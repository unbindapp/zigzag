package mcpserver

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unbindapp/zigzag-api/ent/schema"
	team_repo "github.com/unbindapp/zigzag-api/internal/repositories/team"
	workspace_repo "github.com/unbindapp/zigzag-api/internal/repositories/workspace"
)

func (suite *EndToEndSuite) callError(token, tool string, args map[string]any) (bool, string) {
	result := toolResult{}
	suite.Require().NoError(json.Unmarshal(suite.rpc(token, "tools/call", map[string]any{"name": tool, "arguments": args}), &result))
	suite.Require().Len(result.Content, 1, tool)
	return result.IsError, result.Content[0].Text
}

func (suite *EndToEndSuite) TestAKeyScopedToOneWorkspaceCannotReachAnother() {
	other, err := workspace_repo.NewWorkspaceRepository(suite.DB).Create(suite.Ctx, nil, &workspace_repo.CreateWorkspaceInput{Name: "Globex", Slug: "globex", CreatorID: suite.user.ID})
	suite.Require().NoError(err)
	team, err := team_repo.NewTeamRepository(suite.DB).Create(suite.Ctx, nil, &team_repo.CreateTeamInput{WorkspaceID: other.ID, Name: "Ops", Key: "OPS", CreatorID: suite.user.ID})
	suite.Require().NoError(err)
	issue := suite.DB.Issue.Create().SetWorkspaceID(other.ID).SetTeamID(team.ID).SetNumber(1).SetTitle("Secret").SetStateID(*team.DefaultStateID).SaveX(suite.Ctx)
	comment := suite.DB.Comment.Create().SetIssueID(issue.ID).SetAuthorID(suite.user.ID).SetBody("hidden").SaveX(suite.Ctx)
	label := suite.DB.Label.Create().SetWorkspaceID(other.ID).SetName("Bug").SetColor("#eb5757").SaveX(suite.Ctx)
	status := suite.DB.ProjectStatus.Create().SetWorkspaceID(other.ID).SetName("Backlog").SetType(schema.ProjectStatusBacklog).SetColor("#95a2b3").SaveX(suite.Ctx)
	project := suite.DB.Project.Create().SetWorkspaceID(other.ID).SetName("Moonshot").SetSlug("moonshot-1a2b3c4d").SetStatusID(status.ID).AddTeamIDs(team.ID).SaveX(suite.Ctx)
	milestone := suite.DB.ProjectMilestone.Create().SetProjectID(project.ID).SetName("Beta").SaveX(suite.Ctx)
	now := time.Now()
	cycle := suite.DB.Cycle.Create().SetTeamID(team.ID).SetNumber(1).SetStartsAt(now).SetEndsAt(now.AddDate(0, 0, 14)).SaveX(suite.Ctx)
	document := suite.DB.Document.Create().SetWorkspaceID(other.ID).SetTitle("Plan").SetSlug("plan-1a2b3c4d").SaveX(suite.Ctx)
	initiative := suite.DB.Initiative.Create().SetWorkspaceID(other.ID).SetName("Growth").SetSlug("growth-1a2b3c4d").SaveX(suite.Ctx)
	view := suite.DB.View.Create().SetWorkspaceID(other.ID).SetOwnerID(suite.user.ID).SetName("Mine").SetKind(schema.ViewKindIssue).SetFilters(map[string]any{}).SetDisplay(schema.ViewDisplay{Layout: "list"}).SaveX(suite.Ctx)
	template := suite.DB.Template.Create().SetWorkspaceID(other.ID).SetName("Bug report").SetKind(schema.TemplateKindIssue).SetData(schema.TemplateData{}).SaveX(suite.Ctx)
	webhook := suite.DB.Webhook.Create().SetWorkspaceID(other.ID).SetURL("https://example.com/hook").SetSecret("whsec_x").SetEvents([]string{"issue.created"}).SetLabel("Hook").SaveX(suite.Ctx)

	refused := map[string]map[string]any{
		"get-workspace":           {"workspace_id": other.ID},
		"list-teams":              {"workspace_id": other.ID},
		"list-issues":             {"workspace_id": other.ID},
		"search-issues":           {"workspace_id": other.ID, "query": "Secret"},
		"search":                  {"workspace_id": other.ID, "query": "Secret"},
		"get-team":                {"team_id": team.ID},
		"update-team":             {"team_id": team.ID, "name": "Mine now"},
		"list-workflow-states":    {"team_id": team.ID},
		"get-issue":               {"issue_id": issue.ID},
		"update-issue":            {"issue_id": issue.ID, "title": "Mine now"},
		"delete-issue":            {"issue_id": issue.ID},
		"trash-issue":             {"issue_id": issue.ID},
		"list-issue-history":      {"issue_id": issue.ID},
		"list-comments":           {"issue_id": issue.ID},
		"create-comment":          {"issue_id": issue.ID, "body": "hello"},
		"update-comment":          {"comment_id": comment.ID, "body": "edited"},
		"add-reaction":            {"emoji": "👍", "comment_id": comment.ID},
		"subscribe-issue":         {"issue_id": issue.ID},
		"create-issue":            {"team_id": team.ID, "title": "Planted"},
		"create-issue-relation":   {"issue_id": issue.ID, "related_issue_id": uuid.New(), "type": "related"},
		"list-labels":             {"workspace_id": other.ID},
		"update-label":            {"label_id": label.ID, "name": "Mine now"},
		"delete-label":            {"label_id": label.ID},
		"get-project":             {"project_id": project.ID},
		"update-project":          {"project_id": project.ID, "name": "Mine now"},
		"delete-project":          {"project_id": project.ID},
		"list-milestones":         {"project_id": project.ID},
		"update-milestone":        {"milestone_id": milestone.ID, "name": "Mine now"},
		"get-cycle":               {"cycle_id": cycle.ID},
		"delete-cycle":            {"cycle_id": cycle.ID},
		"get-document":            {"document_id": document.ID},
		"update-document":         {"document_id": document.ID, "title": "Mine now"},
		"get-initiative":          {"initiative_id": initiative.ID},
		"update-initiative":       {"initiative_id": initiative.ID, "name": "Mine now"},
		"get-view":                {"view_id": view.ID},
		"delete-view":             {"view_id": view.ID},
		"get-template":            {"template_id": template.ID},
		"delete-template":         {"template_id": template.ID},
		"list-webhooks":           {"workspace_id": other.ID},
		"delete-webhook":          {"webhook_id": webhook.ID},
		"list-notifications":      {"workspace_id": other.ID},
		"list-favorites":          {"workspace_id": other.ID},
		"add-favorite":            {"workspace_id": other.ID, "kind": "issue", "entity_id": issue.ID},
		"list-workspace-members":  {"workspace_id": other.ID},
		"list-project-statuses":   {"workspace_id": other.ID},
		"list-imports":            {"workspace_id": other.ID},
		"get-notification-counts": {"workspace_id": other.ID},
	}
	for tool, args := range refused {
		failed, text := suite.callError(suite.admin, tool, args)
		suite.True(failed, "%s reached the other workspace: %s", tool, text)
		suite.True(strings.Contains(text, `"status":404`) || strings.Contains(text, `"status":403`), "%s answered %s", tool, text)
	}

	suite.Equal("Secret", suite.DB.Issue.GetX(suite.Ctx, issue.ID).Title, "nothing in the other workspace changed")
	suite.Equal("Ops", suite.DB.Team.GetX(suite.Ctx, team.ID).Name)
	suite.Equal(1, suite.DB.Issue.Query().CountX(suite.Ctx), "no issue was planted in the other team")
	suite.Equal(1, suite.DB.Webhook.Query().CountX(suite.Ctx))
}
