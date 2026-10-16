package user_handler

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
		OperationID: "whoami",
		Summary:     "Who Am I",
		Description: "Get the authenticated user, whether they are a system admin, and every workspace they can see with their role. For an API key or a connected app it also returns the role and workspaces the credential is limited to. Call this first.",
		Path:        "/whoami",
		Method:      http.MethodGet,
	}, handlers.Whoami, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-me",
		Summary:     "Update Me",
		Description: "Change your own name, display name, time zone or avatar.",
		Path:        "/me",
		Method:      http.MethodPut,
	}, handlers.UpdateMe, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "change-password",
		Summary:     "Change Password",
		Description: "Change your own password. Requires the current password.",
		Path:        "/password",
		Method:      http.MethodPut,
	}, handlers.ChangePassword, oapi.SessionOnly, oapi.Confirm)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-notification-settings",
		Summary:     "Get Notification Settings",
		Description: "Get which notifications you receive by email.",
		Path:        "/notification-settings",
		Method:      http.MethodGet,
	}, handlers.GetNotificationSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-notification-settings",
		Summary:     "Update Notification Settings",
		Description: "Change which notifications you receive by email. Omitted fields keep their value.",
		Path:        "/notification-settings",
		Method:      http.MethodPut,
	}, handlers.UpdateNotificationSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-users",
		Summary:     "List Users",
		Description: "List users. System admins see every user; everyone else sees the members of their workspaces.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.ListUsers, oapi.SessionOnly)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-user",
		Summary:     "Create User",
		Description: "Create a user account. System admins only.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.CreateUser, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-user",
		Summary:     "Update User",
		Description: "Change another user's profile, system admin flag, active flag or password. System admins only.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.UpdateUser, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-user",
		Summary:     "Delete User",
		Description: "Delete a user account. System admins only.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteUser, oapi.SessionOnly)
}
