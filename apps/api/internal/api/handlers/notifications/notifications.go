package notifications_handler

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
		OperationID: "list-notifications",
		Summary:     "List Notifications",
		Description: "List your notifications in a workspace, newest first, with the actor, issue and project they point to. filter picks inbox (default), unread, read, archived, snoozed or all. Paginated: pass next_cursor back as cursor.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-notification-counts",
		Summary:     "Get Notification Counts",
		Description: "Count your unread inbox notifications in a workspace. Archived and snoozed notifications are not counted.",
		Path:        "/counts",
		Method:      http.MethodGet,
	}, handlers.Counts, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "mark-notification-read",
		Summary:     "Mark Notifications Read",
		Description: "Mark some of your notifications read by notification_ids. Fails as a whole when one id is not yours.",
		Path:        "/read",
		Method:      http.MethodPost,
	}, handlers.MarkRead, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "mark-notification-unread",
		Summary:     "Mark Notifications Unread",
		Description: "Mark some of your notifications unread again by notification_ids.",
		Path:        "/unread",
		Method:      http.MethodPost,
	}, handlers.MarkUnread, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "archive-notification",
		Summary:     "Archive Notifications",
		Description: "Archive some of your notifications by notification_ids. They leave the inbox, count as read and show with filter archived.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unarchive-notification",
		Summary:     "Unarchive Notifications",
		Description: "Move archived notifications back to the inbox by notification_ids.",
		Path:        "/unarchive",
		Method:      http.MethodPost,
	}, handlers.Unarchive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "snooze-notification",
		Summary:     "Snooze Notifications",
		Description: "Hide some of your notifications until a future time. They come back to the inbox as unread when until passes.",
		Path:        "/snooze",
		Method:      http.MethodPost,
	}, handlers.Snooze, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "mark-all-notifications-read",
		Summary:     "Mark All Notifications Read",
		Description: "Mark every unread notification in your inbox of a workspace read.",
		Path:        "/read-all",
		Method:      http.MethodPost,
	}, handlers.MarkAllRead, oapi.MCP)
}
