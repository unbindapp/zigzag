package issue_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/errdefs"
	"github.com/zigzagapp-org/zigzag-api/internal/integrations/github"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories"
	issue_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/issue"
	permissions_repo "github.com/zigzagapp-org/zigzag-api/internal/repositories/permissions"
)

func (suite *IssueServiceSuite) githubLink(state *ent.WorkflowState) *GithubLink {
	return &GithubLink{
		IssueID: suite.issue.ID,
		Label:   "Pull request #42 opened",
		Attachment: issue_repo.LinkedAttachmentInput{
			Title: "#42 Fix login", URL: "https://github.com/acme/api/pull/42", SourceType: "github",
			Metadata: map[string]any{"repo": "acme/api", "number": 42, "state": "open"},
		},
		ReplaceURL: "https://github.com/acme/api/tree/eng-7",
		State:      state,
		MoveFrom:   []schema.StateType{schema.StateTypeTriage, schema.StateTypeBacklog, schema.StateTypeUnstarted},
	}
}

func (suite *IssueServiceSuite) TestGithubLinkStartsTheIssueAndJoinsTheActiveCycle() {
	suite.loads(suite.issue)
	suite.expectRespond()
	cycleID := uuid.New()
	suite.MockIssueRepo.EXPECT().AttachmentsByURL(mock.Anything, mock.Anything, suite.issue.ID, []string{"https://github.com/acme/api/pull/42", "https://github.com/acme/api/tree/eng-7"}).Return(nil, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateAttachment(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateAttachmentInput) bool {
		return input.CreatorID == nil && input.SourceType == "github" && input.Metadata["state"] == "open"
	})).Return(&ent.Attachment{ID: uuid.New()}, nil).Once()
	suite.MockCycleRepo.EXPECT().ActiveCycleIDs(mock.Anything, []uuid.UUID{suite.team.ID}, suite.now).Return(map[uuid.UUID]uuid.UUID{suite.team.ID: cycleID}, nil).Once()
	suite.MockCycleRepo.EXPECT().GetByID(mock.Anything, cycleID).Return(&ent.Cycle{ID: cycleID, Number: 4}, nil).Once()
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.MatchedBy(func(fields *issue_repo.IssueFields) bool {
		return fields.StateID == suite.states["In Progress"].ID && fields.StartedAt != nil && fields.CycleID != nil && *fields.CycleID == cycleID
	})).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, (*uuid.UUID)(nil), schema.HistorySourceSystem, mock.MatchedBy(suite.changes("github", "state", "cycle"))).Return(nil).Once()

	suite.Require().NoError(suite.service.ApplyGithubLink(suite.Ctx, suite.githubLink(suite.states["In Progress"])))
	suite.Equal([]string{"issue.updated"}, suite.emitted)
}

func (suite *IssueServiceSuite) TestGithubLinkLeavesStatesOutsideMoveFrom() {
	started := suite.newIssue(8, suite.states["In Progress"])
	suite.issue = started
	suite.loads(started)
	existing := &ent.Attachment{ID: uuid.New(), IssueID: started.ID, Title: "#42 Fix login", URL: "https://github.com/acme/api/pull/42", SourceType: "github", Metadata: map[string]any{"repo": "acme/api", "number": float64(42), "state": "open"}}
	suite.MockIssueRepo.EXPECT().AttachmentsByURL(mock.Anything, mock.Anything, started.ID, mock.Anything).Return([]*ent.Attachment{existing}, nil).Once()

	suite.Require().NoError(suite.service.ApplyGithubLink(suite.Ctx, suite.githubLink(suite.states["Done"])))
	suite.Empty(suite.emitted, "an unchanged link on a started issue records nothing")
}

func (suite *IssueServiceSuite) TestCreateFromGithubNeedsTheSystemCaller() {
	_, err := suite.service.CreateFromGithub(suite.Ctx, uuid.Nil, &CreateIssueInput{TeamID: suite.team.ID, Title: "x"}, nil)
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *IssueServiceSuite) TestCreateFromGithubWithoutAnAuthorLinksInTheSameTransaction() {
	ctx := github.WithInbound(permissions_repo.WithSystemCaller(suite.Ctx), "Synced from GitHub #12 by @octocat", "https://github.com/acme/api/issues/12")
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, suite.team.ID).Return(suite.team, nil)
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, uuid.Nil, mock.Anything).Return(nil).Maybe()
	suite.MockIssueRepo.EXPECT().TopSortOrder(mock.Anything, mock.Anything, suite.states["Todo"].ID).Return(0, nil).Once()
	suite.MockTeamRepo.EXPECT().NextIssueNumber(mock.Anything, mock.Anything, suite.team.ID).Return(7, nil).Once()
	suite.MockIssueRepo.EXPECT().Create(mock.Anything, mock.Anything, mock.MatchedBy(func(input *issue_repo.CreateIssueInput) bool {
		return input.CreatorID == nil && len(input.SubscriberIDs) == 0
	})).Return(suite.issue, nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, (*uuid.UUID)(nil), schema.HistorySourceSystem, mock.MatchedBy(func(changes []schema.FieldChange) bool {
		return suite.changes("created", "github")(changes) && changes[1].ToLabel == "Synced from GitHub #12 by @octocat" && changes[1].To == "https://github.com/acme/api/issues/12"
	})).Return(nil).Once()
	suite.loads(suite.issue)
	suite.expectRespond()

	linked := uuid.Nil
	_, err := suite.service.CreateFromGithub(ctx, uuid.Nil, &CreateIssueInput{TeamID: suite.team.ID, Title: "Logs stop streaming", StateID: &suite.states["Todo"].ID}, func(_ context.Context, tx repository.TxInterface, issueID uuid.UUID) error {
		suite.NotNil(tx)
		linked = issueID
		return nil
	})
	suite.Require().NoError(err)
	suite.Equal(suite.issue.ID, linked)
	suite.Equal([]string{"issue.created"}, suite.emitted)
	suite.Empty(suite.events, "nobody to notify without an author, assignee or mention")
}

func (suite *IssueServiceSuite) TestUpdatesHandTheirChangesToGithubSync() {
	suite.loads(suite.issue)
	suite.roles(schema.WorkspaceRoleMember)
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{}).Return([]*ent.Label{}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, suite.issue.ID, []uuid.UUID{}, []uuid.UUID{}).Return(nil)
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, &suite.userID, schema.HistorySourceUser, mock.MatchedBy(suite.changes("title"))).Return(nil).Once()
	suite.expectRespond()

	_, err := suite.service.Update(suite.Ctx, suite.userID, &UpdateIssueInput{IssueID: suite.issue.ID, Title: new("Logs stop streaming on Firefox")})
	suite.Require().NoError(err)
	suite.githubSync.AssertCalled(suite.T(), "AfterIssueUpdate", mock.Anything, suite.issue.ID, mock.MatchedBy(suite.changes("title")))
}

func (suite *IssueServiceSuite) TestInboundUpdatesRecordWhereTheyCameFrom() {
	ctx := github.WithInbound(permissions_repo.WithSystemCaller(suite.Ctx), "Edited on GitHub by @octocat", "https://github.com/acme/api/issues/12")
	suite.loads(suite.issue)
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, uuid.Nil, mock.Anything).Return(nil).Maybe()
	suite.MockIssueRepo.EXPECT().LabelsByIDs(mock.Anything, []uuid.UUID{}).Return([]*ent.Label{}, nil)
	suite.MockIssueRepo.EXPECT().Update(mock.Anything, mock.Anything, suite.issue.ID, mock.Anything).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().SetLabels(mock.Anything, mock.Anything, suite.issue.ID, []uuid.UUID{}, []uuid.UUID{}).Return(nil)
	suite.MockIssueRepo.EXPECT().Subscribe(mock.Anything, mock.Anything, suite.issue.ID).Return(nil).Once()
	suite.MockIssueRepo.EXPECT().CreateHistory(mock.Anything, mock.Anything, suite.issue.ID, (*uuid.UUID)(nil), schema.HistorySourceSystem, mock.MatchedBy(func(changes []schema.FieldChange) bool {
		return suite.changes("title", "github")(changes) && changes[1].ToLabel == "Edited on GitHub by @octocat"
	})).Return(nil).Once()
	suite.expectRespond()

	_, err := suite.service.Update(ctx, uuid.Nil, &UpdateIssueInput{IssueID: suite.issue.ID, Title: new("Logs stop streaming on Firefox")})
	suite.Require().NoError(err)
	suite.Equal([]string{"issue.updated"}, suite.emitted)
}
