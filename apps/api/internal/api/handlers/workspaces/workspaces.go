package workspaces_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
)

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-workspaces",
		Summary:     "List Workspaces",
		Description: "List the workspaces you can see with your role and member and team counts.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-workspace",
		Summary:     "Get Workspace",
		Description: "Get one workspace by workspace_id or slug, with your role and member and team counts.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-workspace",
		Summary:     "Create Workspace",
		Description: "Create a workspace with a name and a unique slug. You become its admin. Default project statuses are created with it. Returns the workspace.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-workspace",
		Summary:     "Update Workspace",
		Description: "Change a workspace's name, slug, icon or color. Returns the workspace. Needs admin.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-workspace",
		Summary:     "Delete Workspace",
		Description: "Delete a workspace with all its teams, issues, projects and documents. Needs admin and the slug repeated as confirm_slug.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-workspace-members",
		Summary:     "List Workspace Members",
		Description: "List a workspace's members with their role, join date and the ids of the teams they are on.",
		Path:        "/members/list",
		Method:      http.MethodGet,
	}, handlers.ListMembers, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "add-workspace-member",
		Summary:     "Add Workspace Member",
		Description: "Add a person by email. An existing account joins at once; otherwise an invitation is created and emailed. Needs admin.",
		Path:        "/members/add",
		Method:      http.MethodPost,
	}, handlers.AddMember, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-workspace-member",
		Summary:     "Update Workspace Member",
		Description: "Change a member's role. A workspace keeps at least one admin. Needs admin.",
		Path:        "/members/update",
		Method:      http.MethodPut,
	}, handlers.UpdateMember, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "remove-workspace-member",
		Summary:     "Remove Workspace Member",
		Description: "Remove a member from the workspace and its teams. Admins remove anyone; anyone may remove themselves to leave.",
		Path:        "/members/remove",
		Method:      http.MethodDelete,
	}, handlers.RemoveMember, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-invitations",
		Summary:     "List Invitations",
		Description: "List pending invitations of a workspace. Links are not included. Needs admin.",
		Path:        "/invitations/list",
		Method:      http.MethodGet,
	}, handlers.ListInvitations, oapi.SessionOnly)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-invitation",
		Summary:     "Create Invitation",
		Description: "Invite an email address with a role and optional teams. Returns the one-time link and whether it was emailed. Needs admin.",
		Path:        "/invitations/create",
		Method:      http.MethodPost,
	}, handlers.CreateInvitation, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "revoke-invitation",
		Summary:     "Revoke Invitation",
		Description: "Revoke a pending invitation so its link stops working. Needs admin.",
		Path:        "/invitations/revoke",
		Method:      http.MethodDelete,
	}, handlers.RevokeInvitation, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "resend-invitation",
		Summary:     "Resend Invitation",
		Description: "Issue a new link for a pending invitation, email it again and extend its expiry. The old link stops working. Needs admin.",
		Path:        "/invitations/resend",
		Method:      http.MethodPost,
	}, handlers.ResendInvitation, oapi.SessionOnly)
}
