package webhook_repo

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/zigzagapp-org/zigzag-api/ent"
	"github.com/zigzagapp-org/zigzag-api/ent/schema"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
	repository "github.com/zigzagapp-org/zigzag-api/internal/repositories/repositorytest"
)

type WebhookRepositorySuite struct {
	repository.RepositoryBaseSuite
	repo      *WebhookRepository
	workspace *ent.Workspace
	team      *ent.Team
}

func (suite *WebhookRepositorySuite) SetupTest() {
	suite.RepositoryBaseSuite.SetupTest()
	suite.repo = NewWebhookRepository(suite.DB)
	suite.workspace = suite.DB.Workspace.Create().SetName("Acme").SetSlug("acme").SaveX(suite.Ctx)
	suite.team = suite.DB.Team.Create().SetWorkspaceID(suite.workspace.ID).SetName("Engineering").SetKey("ENG").SaveX(suite.Ctx)
}

func (suite *WebhookRepositorySuite) create(enabled bool, teamID *uuid.UUID) *ent.Webhook {
	hook, err := suite.repo.Create(suite.Ctx, nil, &CreateWebhookInput{WorkspaceID: suite.workspace.ID, TeamID: teamID, URL: "https://hooks.example.com", Secret: "whsec_x", Enabled: enabled, Events: []string{"issue.created"}, Label: "Bridge"})
	suite.Require().NoError(err)
	return hook
}

func (suite *WebhookRepositorySuite) TestListEnabledAndUpdate() {
	on := suite.create(true, &suite.team.ID)
	suite.create(false, nil)

	all, err := suite.repo.List(suite.Ctx, suite.workspace.ID)
	suite.Require().NoError(err)
	suite.Len(all, 2)
	enabled, err := suite.repo.ListEnabled(suite.Ctx, suite.workspace.ID)
	suite.Require().NoError(err)
	suite.Require().Len(enabled, 1)
	suite.Equal(on.ID, enabled[0].ID)

	updated, err := suite.repo.Update(suite.Ctx, on.ID, &UpdateWebhookInput{Label: utils.ToPtr("Renamed"), Events: []string{"comment.created"}, Enabled: utils.ToPtr(false), ClearTeam: true})
	suite.Require().NoError(err)
	suite.Equal("Renamed", updated.Label)
	suite.Equal([]string{"comment.created"}, updated.Events)
	suite.False(updated.Enabled)
	suite.Nil(updated.TeamID)
	suite.Equal("https://hooks.example.com", updated.URL)

	suite.Require().NoError(suite.repo.Delete(suite.Ctx, on.ID))
	_, err = suite.repo.GetByID(suite.Ctx, on.ID)
	suite.True(ent.IsNotFound(err))
}

func (suite *WebhookRepositorySuite) TestTeamOfIssue() {
	state := suite.DB.WorkflowState.Create().SetTeamID(suite.team.ID).SetName("Todo").SetType(schema.StateTypeUnstarted).SetColor("#e2e2e2").SetPosition(0).SaveX(suite.Ctx)
	issue := suite.DB.Issue.Create().SetWorkspaceID(suite.workspace.ID).SetTeamID(suite.team.ID).SetNumber(1).SetTitle("Issue").SetStateID(state.ID).SaveX(suite.Ctx)
	teamID, err := suite.repo.TeamOfIssue(suite.Ctx, issue.ID)
	suite.Require().NoError(err)
	suite.Equal(suite.team.ID, teamID)
}

func TestWebhookRepositorySuite(t *testing.T) {
	suite.Run(t, new(WebhookRepositorySuite))
}
