package teams_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-teams",
		Summary:     "List Teams",
		Description: "List the teams of a workspace with their workflow states, member counts and whether you are a member. Private teams show only to their members and admins.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-team",
		Summary:     "Get Team",
		Description: "Get a team by team_id, or by workspace_id and key, with its workflow states and settings.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-team",
		Summary:     "Create Team",
		Description: "Create a team with a name and a key used as issue prefix. Default workflow states are created and you join as owner. Needs admin.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-team",
		Summary:     "Update Team",
		Description: "Change a team's name, key, settings, cycles, estimates, triage or defaults. Omitted fields keep their value. Returns the team. Needs admin.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-team",
		Summary:     "Delete Team",
		Description: "Delete a team with all its issues, workflow states and cycles. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-team-members",
		Summary:     "List Team Members",
		Description: "List the members of a team.",
		Path:        "/members/list",
		Method:      http.MethodGet,
	}, handlers.ListMembers, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "add-team-member",
		Summary:     "Add Team Member",
		Description: "Add a workspace member to a team. Admins add anyone; members may add themselves to a team that is not private. Returns the membership.",
		Path:        "/members/add",
		Method:      http.MethodPost,
	}, handlers.AddMember, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "remove-team-member",
		Summary:     "Remove Team Member",
		Description: "Remove someone from a team. Admins remove anyone; anyone may remove themselves.",
		Path:        "/members/remove",
		Method:      http.MethodDelete,
	}, handlers.RemoveMember, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-workflow-states",
		Summary:     "List Workflow States",
		Description: "List a team's workflow states, ordered by type (triage, backlog, unstarted, started, completed, canceled) then position.",
		Path:        "/states/list",
		Method:      http.MethodGet,
	}, handlers.ListStates, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-workflow-state",
		Summary:     "Create Workflow State",
		Description: "Add a workflow state to a team from team_id, name, type and color. Returns the state. Needs admin.",
		Path:        "/states/create",
		Method:      http.MethodPost,
	}, handlers.CreateState, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-workflow-state",
		Summary:     "Update Workflow State",
		Description: "Rename, recolor, describe or reorder a workflow state. Its type cannot change. Returns the state. Needs admin.",
		Path:        "/states/update",
		Method:      http.MethodPut,
	}, handlers.UpdateState, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-workflow-state",
		Summary:     "Delete Workflow State",
		Description: "Delete a workflow state after moving its issues to another state of the same team. Needs admin.",
		Path:        "/states/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteState, oapi.MCP)
}
