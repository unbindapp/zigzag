package github_service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
	"github.com/unbindapp/zigzag-api/internal/common/errdefs"
	"github.com/unbindapp/zigzag-api/internal/integrations/github"
	"github.com/unbindapp/zigzag-api/internal/models"
	github_repo "github.com/unbindapp/zigzag-api/internal/repositories/github"
	permissions_repo "github.com/unbindapp/zigzag-api/internal/repositories/permissions"
	comment_service "github.com/unbindapp/zigzag-api/internal/services/comment"
	issue_service "github.com/unbindapp/zigzag-api/internal/services/issue"
	project_service "github.com/unbindapp/zigzag-api/internal/services/project"
)

func (suite *GithubSyncSuite) admin(result error) {
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, []permissions_repo.PermissionCheck{{Role: schema.WorkspaceRoleAdmin, WorkspaceID: suite.workspaceID}}).Return(result).Maybe()
}

func (suite *GithubSyncSuite) TestTokenNeedsAdminAndIsStoredSealed() {
	suite.admin(errdefs.ErrUnauthorized)
	_, err := suite.service.UpdateIntegration(suite.Ctx, suite.userID, &UpdateGithubIntegrationInput{WorkspaceID: suite.workspaceID, Token: syncToken})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)
}

func (suite *GithubSyncSuite) TestTokenIsValidatedSealedAndCleared() {
	suite.admin(nil)
	for _, token := range []string{"short", "github_pat_ with spaces inside the token"} {
		_, err := suite.service.UpdateIntegration(suite.Ctx, suite.userID, &UpdateGithubIntegrationInput{WorkspaceID: suite.workspaceID, Token: token})
		suite.ErrorIs(err, errdefs.ErrInvalidInput, token)
	}

	suite.MockGithubRepo.EXPECT().UpdateToken(mock.Anything, suite.integration.ID, mock.Anything, "github_pat_11AB").RunAndReturn(func(_ context.Context, _ uuid.UUID, sealed []byte, _ string) (*ent.GithubIntegration, error) {
		opened, err := suite.encryptor.Decrypt(sealed)
		suite.Require().NoError(err)
		suite.Equal(syncToken, string(opened))
		return suite.integration, nil
	}).Once()
	updated, err := suite.service.UpdateIntegration(suite.Ctx, suite.userID, &UpdateGithubIntegrationInput{WorkspaceID: suite.workspaceID, Token: "  " + syncToken + " "})
	suite.Require().NoError(err)
	suite.True(updated.HasToken)

	suite.MockGithubRepo.EXPECT().UpdateToken(mock.Anything, suite.integration.ID, []byte(nil), "").Return(&ent.GithubIntegration{ID: suite.integration.ID, WorkspaceID: suite.workspaceID}, nil).Once()
	cleared, err := suite.service.UpdateIntegration(suite.Ctx, suite.userID, &UpdateGithubIntegrationInput{WorkspaceID: suite.workspaceID, Token: ""})
	suite.Require().NoError(err)
	suite.False(cleared.HasToken)
}

func (suite *GithubSyncSuite) TestTokenPrefixesShowOnlyTheKind() {
	for token, prefix := range map[string]string{
		syncToken: "github_pat_11AB",
		"ghp_0123456789abcdefghijklmnopqrstuvwxyz": "ghp_0123",
		"0123456789abcdefghijklmnop":               "0123",
	} {
		suite.Equal(prefix, tokenPrefix(token))
	}
}

func (suite *GithubSyncSuite) TestListingMasksOtherWorkspacesAsNotFound() {
	suite.admin(errdefs.ErrUnauthorized)
	_, err := suite.service.ListRepoSyncs(suite.Ctx, suite.userID, &ListGithubRepoSyncsInput{WorkspaceID: suite.workspaceID})
	suite.True(errdefs.IsNotFound(err))
}

func (suite *GithubSyncSuite) TestListingCountsLinkedIssues() {
	suite.admin(nil)
	suite.MockGithubRepo.EXPECT().ListRepoSyncs(mock.Anything, suite.integration.ID).Return([]*ent.GithubRepoSync{suite.repoSync}, nil).Once()
	suite.MockGithubRepo.EXPECT().LinkedIssueCounts(mock.Anything, []uuid.UUID{suite.repoSync.ID}).Return(map[uuid.UUID]int{suite.repoSync.ID: 4}, nil).Once()
	syncs, err := suite.service.ListRepoSyncs(suite.Ctx, suite.userID, &ListGithubRepoSyncsInput{WorkspaceID: suite.workspaceID})
	suite.Require().NoError(err)
	suite.Require().Len(syncs, 1)
	suite.Equal(4, syncs[0].LinkedIssues)
	suite.Equal("UNB", syncs[0].Team.Key)
	suite.Equal(suite.workspaceID, syncs[0].WorkspaceID)
}

func (suite *GithubSyncSuite) TestAddingValidatesNameTeamAndAccess() {
	suite.admin(nil)
	add := func(fullName string, teamID uuid.UUID) error {
		_, err := suite.service.AddRepoSync(suite.Ctx, suite.userID, &AddGithubRepoSyncInput{WorkspaceID: suite.workspaceID, FullName: fullName, TeamID: teamID})
		return err
	}
	suite.ErrorIs(add("just-a-name", suite.team.ID), errdefs.ErrInvalidInput)

	elsewhere := uuid.New()
	suite.MockTeamRepo.EXPECT().GetByID(mock.Anything, elsewhere).Return(&ent.Team{ID: elsewhere, WorkspaceID: uuid.New()}, nil).Once()
	suite.ErrorIs(add("unbindapp/zigzag", elsewhere), errdefs.ErrInvalidInput)

	suite.MockGithubRepo.EXPECT().RepoSyncNameTaken(mock.Anything, "unbindapp/unbind").Return(true, nil).Once()
	suite.ErrorIs(add("https://github.com/UnbindApp/Unbind.git", suite.team.ID), errdefs.ErrConflict)

	suite.MockGithubRepo.EXPECT().RepoSyncNameTaken(mock.Anything, "unbindapp/private").Return(false, nil).Once()
	suite.client.EXPECT().GetRepository(mock.Anything, syncToken, "unbindapp/private").Return(nil, &github.APIError{Status: 404, Message: "Not Found"}).Once()
	err := add("unbindapp/private", suite.team.ID)
	suite.ErrorIs(err, errdefs.ErrInvalidInput)
	suite.Contains(errdefs.DetailMessage(err), "cannot find unbindapp/private")
}

func (suite *GithubSyncSuite) TestAddingCreatesTheProjectOrReusesIt() {
	suite.admin(nil)
	suite.MockGithubRepo.EXPECT().RepoSyncNameTaken(mock.Anything, mock.Anything).Return(false, nil).Twice()
	suite.client.EXPECT().GetRepository(mock.Anything, syncToken, mock.Anything).Return(&github.RepositoryInfo{HasIssues: true}, nil).Twice()
	suite.MockGithubRepo.EXPECT().LinkedIssueCounts(mock.Anything, mock.Anything).Return(map[uuid.UUID]int{}, nil).Twice()

	createdID := uuid.New()
	suite.MockProjectRepo.EXPECT().GetByName(mock.Anything, suite.workspaceID, "unbindapp/zigzag").Return(nil, &ent.NotFoundError{}).Once()
	suite.projects.EXPECT().Create(mock.Anything, suite.userID, &project_service.CreateProjectInput{
		WorkspaceID: suite.workspaceID, Name: "unbindapp/zigzag", TeamIDs: []uuid.UUID{suite.team.ID}, Summary: "Issues synced from GitHub",
	}).Return(&models.ProjectResponse{ID: createdID}, nil).Once()
	suite.MockGithubRepo.EXPECT().CreateRepoSync(mock.Anything, &github_repo.CreateRepoSyncInput{
		IntegrationID: suite.integration.ID, FullName: "unbindapp/zigzag", TeamID: suite.team.ID, ProjectID: createdID, SyncClosed: true,
	}).Return(suite.repoSync, nil).Once()
	_, err := suite.service.AddRepoSync(suite.Ctx, suite.userID, &AddGithubRepoSyncInput{WorkspaceID: suite.workspaceID, FullName: "unbindapp/zigzag", TeamID: suite.team.ID, SyncClosed: true})
	suite.Require().NoError(err)

	otherTeam := &ent.Team{ID: uuid.New()}
	existing := &ent.Project{ID: uuid.New(), Name: "unbindapp/website"}
	existing.Edges.Teams = []*ent.Team{otherTeam}
	suite.MockProjectRepo.EXPECT().GetByName(mock.Anything, suite.workspaceID, "unbindapp/website").Return(existing, nil).Once()
	suite.projects.EXPECT().Update(mock.Anything, suite.userID, &project_service.UpdateProjectInput{ProjectID: existing.ID, TeamIDs: []uuid.UUID{suite.team.ID, otherTeam.ID}}).Return(&models.ProjectResponse{}, nil).Once()
	suite.MockGithubRepo.EXPECT().CreateRepoSync(mock.Anything, mock.MatchedBy(func(input *github_repo.CreateRepoSyncInput) bool {
		return input.ProjectID == existing.ID
	})).Return(suite.repoSync, nil).Once()
	_, err = suite.service.AddRepoSync(suite.Ctx, suite.userID, &AddGithubRepoSyncInput{WorkspaceID: suite.workspaceID, FullName: "unbindapp/website", TeamID: suite.team.ID})
	suite.Require().NoError(err)
}

func (suite *GithubSyncSuite) TestRemovingNeedsAdmin() {
	suite.MockGithubRepo.EXPECT().GetRepoSync(mock.Anything, suite.repoSync.ID).Return(suite.repoSync, nil).Twice()
	suite.MockPermissionsRepo.EXPECT().Check(mock.Anything, suite.userID, mock.Anything).Return(errdefs.ErrUnauthorized).Once()
	_, err := suite.service.RemoveRepoSync(suite.Ctx, suite.userID, &GithubRepoSyncRefInput{RepoSyncID: suite.repoSync.ID})
	suite.ErrorIs(err, errdefs.ErrUnauthorized)

	suite.admin(nil)
	suite.MockGithubRepo.EXPECT().DeleteRepoSync(mock.Anything, suite.repoSync.ID).Return(nil).Once()
	removed, err := suite.service.RemoveRepoSync(suite.Ctx, suite.userID, &GithubRepoSyncRefInput{RepoSyncID: suite.repoSync.ID})
	suite.Require().NoError(err)
	suite.Equal(suite.repoSync.ID, removed)
}

func (suite *GithubSyncSuite) TestBackfillNeedsATokenAndRunsOnce() {
	suite.admin(nil)
	suite.MockGithubRepo.EXPECT().GetRepoSync(mock.Anything, suite.repoSync.ID).Return(suite.repoSync, nil).Maybe()
	suite.integration.TokenEncrypted = nil
	_, err := suite.service.Backfill(suite.Ctx, suite.userID, &BackfillGithubRepoSyncInput{RepoSyncID: suite.repoSync.ID})
	suite.ErrorIs(err, errdefs.ErrInvalidInput)

	sealed, err := suite.encryptor.Encrypt([]byte(syncToken))
	suite.Require().NoError(err)
	suite.integration.TokenEncrypted = sealed
	suite.MockGithubRepo.EXPECT().StartBackfill(mock.Anything, suite.repoSync.ID, true).Return(false, nil).Once()
	_, err = suite.service.Backfill(suite.Ctx, suite.userID, &BackfillGithubRepoSyncInput{RepoSyncID: suite.repoSync.ID, SyncClosed: new(true)})
	suite.ErrorIs(err, errdefs.ErrConflict)
	suite.Empty(suite.jobs)
}

func (suite *GithubSyncSuite) githubIssue(number int, nodeID, updatedAt string, comments int) github.Issue {
	at, err := time.Parse(time.RFC3339, updatedAt)
	suite.Require().NoError(err)
	return github.Issue{
		Number: number, NodeID: nodeID, Title: "Logs stop streaming", Body: "Body", State: "open",
		HTMLURL: "https://github.com/unbindapp/unbind/issues/" + nodeID, User: github.Account{Login: "octocat"},
		Labels: []github.Label{}, Comments: comments, UpdatedAt: at, CreatedAt: at,
	}
}

func (suite *GithubSyncSuite) TestBackfillCountsCreatedUpdatedAndSkipped() {
	suite.admin(nil)
	suite.MockGithubRepo.EXPECT().GetRepoSync(mock.Anything, suite.repoSync.ID).Return(suite.repoSync, nil).Maybe()
	suite.MockGithubRepo.EXPECT().StartBackfill(mock.Anything, suite.repoSync.ID, false).Return(true, nil).Once()
	suite.MockGithubRepo.EXPECT().LinkedIssueCounts(mock.Anything, mock.Anything).Return(map[uuid.UUID]int{}, nil).Once()
	_, err := suite.service.Backfill(suite.Ctx, suite.userID, &BackfillGithubRepoSyncInput{RepoSyncID: suite.repoSync.ID})
	suite.Require().NoError(err)
	suite.Require().Len(suite.jobs, 1)

	pull := suite.githubIssue(1, "PR_1", "2026-10-16T11:00:00Z", 0)
	pull.PullRequest = []byte(`{"url":"x"}`)
	fresh := suite.githubIssue(2, "I_fresh", "2026-10-16T11:00:00Z", 2)
	unchanged := suite.githubIssue(12, issueNodeID, "2026-10-16T10:00:00Z", 1)
	changedLink := &ent.GithubIssueLink{ID: uuid.New(), IssueID: suite.issue.ID, RepoSyncID: suite.repoSync.ID, Number: 13, NodeID: "I_changed", GithubUpdatedAt: time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)}
	changedLink.Edges.RepoSync = suite.repoSync
	changed := suite.githubIssue(13, "I_changed", "2026-10-16T11:30:00Z", 0)
	changed.State, changed.StateReason = "closed", "completed"

	suite.client.EXPECT().ListIssues(mock.Anything, syncToken, repoName, "open", 1).Return([]github.Issue{pull, fresh, unchanged}, 2, nil).Once()
	suite.client.EXPECT().ListIssues(mock.Anything, syncToken, repoName, "open", 2).Return([]github.Issue{changed}, 0, nil).Once()

	freshLink := &ent.GithubIssueLink{ID: uuid.New(), IssueID: uuid.New(), RepoSyncID: suite.repoSync.ID, Number: 2, NodeID: "I_fresh"}
	freshLink.Edges.RepoSync = suite.repoSync
	suite.issues.EXPECT().CreateFromGithub(mock.Anything, uuid.Nil, mock.MatchedBy(func(input *issue_service.CreateIssueInput) bool {
		return input.Title == "Logs stop streaming"
	}), mock.Anything).RunAndReturn(func(context.Context, uuid.UUID, *issue_service.CreateIssueInput, issue_service.LinkInTx) (*models.IssueResponse, error) {
		suite.extraLinks["I_fresh"] = freshLink
		return &models.IssueResponse{}, nil
	}).Once()
	suite.client.EXPECT().ListComments(mock.Anything, syncToken, repoName, 2, 1).Return([]github.Comment{
		{ID: 501, Body: "First", User: github.Account{Login: "octocat"}, CreatedAt: time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)},
		{ID: 502, Body: "**Eve on Zigzag**\n\nAn echo", User: github.Account{Login: "zigzag-bot"}},
	}, 0, nil).Once()
	suite.MockGithubRepo.EXPECT().CommentLinkByGithubID(mock.Anything, int64(501)).Return(nil, &ent.NotFoundError{}).Once()
	suite.comments.EXPECT().CreateFromGithub(mock.Anything, mock.MatchedBy(func(input *comment_service.GithubCommentInput) bool {
		return input.IssueID == freshLink.IssueID && input.CreatedAt.Equal(time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC))
	})).Return(&models.CommentResponse{}, nil).Once()

	suite.extraLinks["I_changed"] = changedLink
	suite.issues.EXPECT().Update(mock.Anything, uuid.Nil, mock.MatchedBy(func(input *issue_service.UpdateIssueInput) bool {
		return input.StateID != nil && *input.StateID == suite.states["Done"].ID && input.Title == nil
	})).RunAndReturn(func(ctx context.Context, _ uuid.UUID, _ *issue_service.UpdateIssueInput) (*models.IssueResponse, error) {
		suite.Equal("Synced from GitHub #13", inboundLabel(ctx))
		return &models.IssueResponse{}, nil
	}).Once()
	suite.MockGithubRepo.EXPECT().RecordInbound(mock.Anything, changedLink.ID, time.Date(2026, 10, 16, 11, 30, 0, 0, time.UTC)).Return(nil).Once()

	suite.MockGithubRepo.EXPECT().RecordBackfillProgress(mock.Anything, suite.repoSync.ID, mock.Anything).Return(nil).Twice()
	suite.MockGithubRepo.EXPECT().FinishBackfill(mock.Anything, suite.repoSync.ID, schema.GithubBackfillCompleted, &schema.GithubBackfillReport{
		IssuesCreated: 1, IssuesUpdated: 1, CommentsCreated: 1, Skipped: 2,
	}, "", suite.now).Return(nil).Once()
	suite.jobs[0]()
}

func (suite *GithubSyncSuite) TestBackfillThatCannotListFails() {
	suite.client.EXPECT().ListIssues(mock.Anything, syncToken, repoName, "all", 1).Return(nil, 0, &github.APIError{Status: 401, Message: "Bad credentials"}).Once()
	suite.MockGithubRepo.EXPECT().FinishBackfill(mock.Anything, suite.repoSync.ID, schema.GithubBackfillFailed, &schema.GithubBackfillReport{}, "GitHub answered 401: Bad credentials", suite.now).Return(nil).Once()
	suite.repoSync.SyncClosed = true
	suite.service.backfill(suite.Ctx, suite.repoSync, syncToken)
}
