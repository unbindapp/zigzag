package issues_handler

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
		OperationID: "list-issues",
		Summary:     "List Issues",
		Description: "List a workspace's issues with filters, in manual order unless order_by says otherwise, 50 per page. Archived and trashed issues are hidden unless asked for. Pass next_cursor back as cursor for the next page; total counts every match.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-issue",
		Summary:     "Get Issue",
		Description: "Get one issue by issue_id or by identifier like ENG-123, with sub-issues, relations, attachments, subscribers and reactions. Identifiers from before a team move still resolve.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-issue",
		Summary:     "Create Issue",
		Description: "Create an issue in a team. Only team_id and title are needed: the state defaults to the team's default, priority to none, and the next number of the team gives the identifier. You, the assignee and mentioned people are subscribed. Needs member.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-issue",
		Summary:     "Update Issue",
		Description: "Change any field of an issue; omitted fields keep their value and the nil uuid clears an id field. Moving to a started, completed or canceled state sets its timestamp, and team_id moves the issue to another team with a new identifier. Every change is recorded in the history. Needs member.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-issue",
		Summary:     "Delete Issue",
		Description: "Delete an issue permanently with its comments, attachments and history; its sub-issues become top level. Prefer trash-issue, which can be undone. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "trash-issue",
		Summary:     "Trash Issue",
		Description: "Move an issue to the trash, hiding it from lists. It is deleted for good 30 days later unless restored with restore-issue. Needs member.",
		Path:        "/trash",
		Method:      http.MethodPost,
	}, handlers.Trash, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "restore-issue",
		Summary:     "Restore Issue",
		Description: "Take an issue out of the trash. Returns the issue. Needs member.",
		Path:        "/restore",
		Method:      http.MethodPost,
	}, handlers.Restore, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "archive-issue",
		Summary:     "Archive Issue",
		Description: "Archive an issue, hiding it from lists unless include_archived is set. Returns the issue. Needs member.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unarchive-issue",
		Summary:     "Unarchive Issue",
		Description: "Bring an archived issue back into lists. Returns the issue. Needs member.",
		Path:        "/unarchive",
		Method:      http.MethodPost,
	}, handlers.Unarchive, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "bulk-update-issues",
		Summary:     "Bulk Update Issues",
		Description: "Apply the same change to up to 250 issues in one transaction: all succeed or none do. Each issue gets its own history entry. Needs member.",
		Path:        "/bulk-update",
		Method:      http.MethodPut,
	}, handlers.BulkUpdate, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-issue-history",
		Summary:     "List Issue History",
		Description: "List the changes made to an issue, oldest first, each with its actor, source and field changes. Paginated with cursor like list-issues.",
		Path:        "/history/list",
		Method:      http.MethodGet,
	}, handlers.ListHistory, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-comments",
		Summary:     "List Comments",
		Description: "List the comment threads of an issue, oldest first, each with its replies and reactions. Paginated by thread with cursor.",
		Path:        "/comments/list",
		Method:      http.MethodGet,
	}, handlers.ListComments, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-comment",
		Summary:     "Create Comment",
		Description: "Comment on an issue in markdown, or reply with parent_id. Mentioned people and the issue's subscribers are notified, and you are subscribed. Needs member.",
		Path:        "/comments/create",
		Method:      http.MethodPost,
	}, handlers.CreateComment, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-comment",
		Summary:     "Update Comment",
		Description: "Replace the body of a comment. Only its author or an admin may.",
		Path:        "/comments/update",
		Method:      http.MethodPut,
	}, handlers.UpdateComment, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-comment",
		Summary:     "Delete Comment",
		Description: "Delete a comment and its replies. Only its author or an admin may.",
		Path:        "/comments/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteComment, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "resolve-comment",
		Summary:     "Resolve Comment",
		Description: "Mark a comment thread as resolved. Needs member.",
		Path:        "/comments/resolve",
		Method:      http.MethodPost,
	}, handlers.ResolveComment, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unresolve-comment",
		Summary:     "Unresolve Comment",
		Description: "Reopen a resolved comment thread. Needs member.",
		Path:        "/comments/unresolve",
		Method:      http.MethodPost,
	}, handlers.UnresolveComment, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-issue-relation",
		Summary:     "Create Issue Relation",
		Description: "Relate two issues of a workspace: issue_id blocks, relates to or duplicates related_issue_id. A duplicate moves to its team's Duplicate state, and the assignee of a blocked issue is notified. Needs member.",
		Path:        "/relations/create",
		Method:      http.MethodPost,
	}, handlers.CreateRelation, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-issue-relation",
		Summary:     "Delete Issue Relation",
		Description: "Remove a relation between two issues by its relation_id from get-issue. Needs member.",
		Path:        "/relations/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteRelation, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-attachment",
		Summary:     "Create Attachment",
		Description: "Attach a link (url) or a file uploaded to the same workspace (upload_id) to an issue. Returns the attachment. Needs member.",
		Path:        "/attachments/create",
		Method:      http.MethodPost,
	}, handlers.CreateAttachment, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-attachment",
		Summary:     "Update Attachment",
		Description: "Change the title or subtitle of an attachment. Returns the attachment. Needs member.",
		Path:        "/attachments/update",
		Method:      http.MethodPut,
	}, handlers.UpdateAttachment, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-attachment",
		Summary:     "Delete Attachment",
		Description: "Remove an attachment from its issue. Needs member.",
		Path:        "/attachments/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteAttachment, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "subscribe-issue",
		Summary:     "Subscribe Issue",
		Description: "Subscribe yourself, or the workspace member in user_id, to an issue's notifications. Needs member.",
		Path:        "/subscribe",
		Method:      http.MethodPost,
	}, handlers.Subscribe, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unsubscribe-issue",
		Summary:     "Unsubscribe Issue",
		Description: "Stop notifications about an issue for yourself, or for the member in user_id. Needs member.",
		Path:        "/unsubscribe",
		Method:      http.MethodPost,
	}, handlers.Unsubscribe, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "add-reaction",
		Summary:     "Add Reaction",
		Description: "React with an emoji to an issue, a comment or a project update; give exactly one of their ids. Returns the target's reactions. Needs member.",
		Path:        "/reactions/add",
		Method:      http.MethodPost,
	}, handlers.AddReaction, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "remove-reaction",
		Summary:     "Remove Reaction",
		Description: "Take back your emoji reaction from an issue, a comment or a project update. Returns the target's reactions. Needs member.",
		Path:        "/reactions/remove",
		Method:      http.MethodDelete,
	}, handlers.RemoveReaction, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "search-issues",
		Summary:     "Search Issues",
		Description: "Find issues of a workspace by fuzzy text, ranking title matches above description matches; an identifier like ENG-123 comes first. Trashed issues are left out.",
		Path:        "/search",
		Method:      http.MethodGet,
	}, handlers.Search, oapi.MCP)
}
