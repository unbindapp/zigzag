package issue_repo

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/unbindapp/zigzag-api/internal/repositories/repositorytest"
)

func TestIssueRepositoryOnPostgres(t *testing.T) {
	if !repositorytest.PostgresConfigured() {
		t.Skip("set " + repositorytest.PostgresURLEnv + " to run the issue repository against Postgres")
	}
	run := new(IssueRepositorySuite)
	run.Postgres = true
	suite.Run(t, run)
}

func (suite *IssueRepositorySuite) searchTitles(query string, teams []uuid.UUID, limit int) []string {
	ids, err := suite.repo.Search(suite.Ctx, suite.workspace.ID, teams, query, limit)
	suite.Require().NoError(err)
	issues, err := suite.repo.GetManyWithEdges(suite.Ctx, nil, ids)
	suite.Require().NoError(err)
	titles := make([]string, len(issues))
	for i, issue := range issues {
		titles[i] = issue.Title
	}
	return titles
}

func (suite *IssueRepositorySuite) TestSearchMatchesEveryWordAndRanksTitlesFirst() {
	repositorytest.SkipWithoutPostgres(&suite.RepositoryBaseSuite)
	suite.create("Login page crashes on submit", nil)
	suite.create("Settings layout", func(in *CreateIssueInput) { in.DescriptionText = "the login form crashes when the password is empty" })
	suite.create("Crash reporter", nil)
	suite.create("Ops login audit", func(in *CreateIssueInput) { in.TeamID = suite.ops.ID })
	trashed := suite.create("Login crashes in trash", nil)
	suite.DB.Issue.UpdateOneID(trashed.ID).SetTrashedAt(time.Now()).ExecX(suite.Ctx)

	both := []uuid.UUID{suite.eng.ID, suite.ops.ID}
	suite.Equal([]string{"Login page crashes on submit", "Settings layout"}, suite.searchTitles("login crashes", both, 10), "every word must match; title hits rank first; trash is left out")
	suite.Equal([]string{"Login page crashes on submit"}, suite.searchTitles("login crashes", both, 1), "limit applies")
	suite.Equal([]string{"Ops login audit"}, suite.searchTitles("audit", both, 10))
	suite.Empty(suite.searchTitles("audit", []uuid.UUID{suite.eng.ID}, 10), "teams outside the filter are left out")
	suite.Equal([]string{"Login page crashes on submit"}, suite.searchTitles("submitt", both, 10), "a misspelled word still matches by trigram word similarity")
	suite.Empty(suite.searchTitles("100%", both, 10), "like wildcards in the query are literal")
	suite.Empty(suite.searchTitles("   ", both, 10))
}

func (suite *IssueRepositorySuite) TestLockWithEdgesHoldsTheRowUntilCommit() {
	repositorytest.SkipWithoutPostgres(&suite.RepositoryBaseSuite)
	created := suite.create("Locked", nil)
	holder, err := suite.DB.Tx(suite.Ctx)
	suite.Require().NoError(err)
	locked, err := suite.repo.LockWithEdges(suite.Ctx, holder, created.ID)
	suite.Require().NoError(err)
	suite.NotNil(locked.Edges.State, "the locked read loads the embedded shapes")

	waiter, err := suite.DB.Tx(suite.Ctx)
	suite.Require().NoError(err)
	defer func() { _ = waiter.Rollback() }()
	blocked, cancel := context.WithTimeout(suite.Ctx, 300*time.Millisecond)
	defer cancel()
	_, err = suite.repo.LockWithEdges(blocked, waiter, created.ID)
	suite.Error(err, "a second writer waits while the first transaction holds the row")
	suite.Require().NoError(holder.Commit())
}
