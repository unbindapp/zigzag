package issue_service

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	issue_repo "github.com/unbindapp/zigzag-api/internal/repositories/issue"
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
