package github_repo

import (
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

func (suite *GithubRepositorySuite) syncedRepo() (*ent.GithubRepoSync, *ent.Issue) {
	integration := suite.create()
	team := suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Eng").SetKey("ENG").SaveX(suite.Ctx)
	state := suite.DB.WorkflowState.Create().SetTeamID(team.ID).SetName("Todo").SetType("unstarted").SetColor("#ffffff").SaveX(suite.Ctx)
	issue := suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(team.ID).SetNumber(1).SetTitle("t").SetStateID(state.ID).SaveX(suite.Ctx)
	status := suite.DB.ProjectStatus.Create().SetWorkspaceID(suite.workspace.ID).SetName("Backlog").SetType("backlog").SetColor("#ffffff").SaveX(suite.Ctx)
	project := suite.DB.Project.Create().SetWorkspaceID(suite.workspace.ID).SetName("unbindapp/zigzag").SetSlug("zigzag").SetStatusID(status.ID).SaveX(suite.Ctx)
	created, err := suite.repo.CreateRepoSync(suite.Ctx, &CreateRepoSyncInput{IntegrationID: integration.ID, FullName: "unbindapp/zigzag", TeamID: team.ID, ProjectID: project.ID})
	suite.Require().NoError(err)
	return created, issue
}

func (suite *GithubRepositorySuite) TestRepoSyncsAreUniqueAndLoadTheirEdges() {
	created, _ := suite.syncedRepo()
	suite.Equal(schema.GithubBackfillNone, created.BackfillStatus)
	suite.Require().NotNil(created.Edges.Team)
	suite.Require().NotNil(created.Edges.Project)
	suite.Require().NotNil(created.Edges.Integration)
	taken, err := suite.repo.RepoSyncNameTaken(suite.Ctx, "unbindapp/zigzag")
	suite.Require().NoError(err)
	suite.True(taken)
	_, err = suite.repo.CreateRepoSync(suite.Ctx, &CreateRepoSyncInput{IntegrationID: created.IntegrationID, FullName: "unbindapp/zigzag", TeamID: created.TeamID, ProjectID: *created.ProjectID})
	suite.True(ent.IsConstraintError(err))
}

func (suite *GithubRepositorySuite) TestOnlyOneBackfillRunsAtATime() {
	created, _ := suite.syncedRepo()
	started, err := suite.repo.StartBackfill(suite.Ctx, created.ID, true)
	suite.Require().NoError(err)
	suite.True(started)
	started, err = suite.repo.StartBackfill(suite.Ctx, created.ID, true)
	suite.Require().NoError(err)
	suite.False(started)

	suite.Require().NoError(suite.repo.FailInterruptedBackfills(suite.Ctx))
	failed, err := suite.repo.GetRepoSync(suite.Ctx, created.ID)
	suite.Require().NoError(err)
	suite.Equal(schema.GithubBackfillFailed, failed.BackfillStatus)
	suite.True(failed.SyncClosed)

	report := &schema.GithubBackfillReport{IssuesCreated: 3, Skipped: 1}
	at := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	started, err = suite.repo.StartBackfill(suite.Ctx, created.ID, false)
	suite.Require().NoError(err)
	suite.True(started)
	suite.Require().NoError(suite.repo.FinishBackfill(suite.Ctx, created.ID, schema.GithubBackfillCompleted, report, "", at))
	finished, err := suite.repo.GetRepoSync(suite.Ctx, created.ID)
	suite.Require().NoError(err)
	suite.Equal(schema.GithubBackfillCompleted, finished.BackfillStatus)
	suite.Equal(report, finished.BackfillReport)
	suite.Empty(finished.LastError)
	suite.True(finished.LastSyncedAt.Equal(at))
}

func (suite *GithubRepositorySuite) TestIssueLinksOnlyMoveForwardAndCount() {
	created, issue := suite.syncedRepo()
	first := time.Date(2026, 10, 16, 10, 0, 0, 0, time.UTC)
	link, err := suite.repo.CreateIssueLink(suite.Ctx, nil, &IssueLinkInput{IssueID: issue.ID, RepoSyncID: created.ID, Number: 12, NodeID: "I_12", HTMLURL: "https://github.com/unbindapp/zigzag/issues/12", GithubUpdatedAt: first})
	suite.Require().NoError(err)
	suite.Require().NoError(suite.repo.RecordInbound(suite.Ctx, link.ID, first.Add(-time.Hour)))
	found, err := suite.repo.IssueLinkByNodeID(suite.Ctx, "I_12")
	suite.Require().NoError(err)
	suite.True(found.GithubUpdatedAt.Equal(first), "an older delivery does not move the link back")
	suite.Require().NotNil(found.Edges.RepoSync)
	suite.Require().NotNil(found.Edges.RepoSync.Edges.Integration)

	suite.Require().NoError(suite.repo.RecordInbound(suite.Ctx, link.ID, first.Add(time.Hour)))
	found, err = suite.repo.IssueLinkByIssue(suite.Ctx, issue.ID)
	suite.Require().NoError(err)
	suite.True(found.GithubUpdatedAt.Equal(first.Add(time.Hour)))

	counts, err := suite.repo.LinkedIssueCounts(suite.Ctx, []uuid.UUID{created.ID})
	suite.Require().NoError(err)
	suite.Equal(map[uuid.UUID]int{created.ID: 1}, counts)

	comment := suite.DB.Comment.Create().SetIssueID(issue.ID).SetBody("x").SaveX(suite.Ctx)
	suite.Require().NoError(suite.repo.CreateCommentLink(suite.Ctx, nil, comment.ID, link.ID, 9001))
	suite.DB.Comment.DeleteOneID(comment.ID).ExecX(suite.Ctx)
	kept, err := suite.repo.CommentLinkByComment(suite.Ctx, comment.ID)
	suite.Require().NoError(err, "the link outlives the comment until GitHub deletes its copy")
	suite.Equal(int64(9001), kept.GithubCommentID)

	suite.Require().NoError(suite.repo.DeleteRepoSync(suite.Ctx, created.ID))
	suite.Zero(suite.DB.GithubIssueLink.Query().CountX(suite.Ctx))
	suite.Zero(suite.DB.GithubCommentLink.Query().CountX(suite.Ctx))
	suite.Equal(1, suite.DB.Issue.Query().CountX(suite.Ctx), "issues stay when a repository stops syncing")
}
