package projects_handler

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
		OperationID: "list-projects",
		Summary:     "List Projects",
		Description: "List a workspace's projects with status, lead, members, teams, milestones and progress. Filter by team, status, lead, member, initiative, label or text; archived and trashed projects are hidden unless include_archived or include_trashed is set. Not paginated.",
		Path:        "/list",
		Method:      http.MethodGet,
	}, handlers.List, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-project",
		Summary:     "Get Project",
		Description: "Get a project by project_id, or by workspace_id and slug, with teams, members, milestones, links, progress, issue counts and the latest status update.",
		Path:        "/get",
		Method:      http.MethodGet,
	}, handlers.Get, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-project",
		Summary:     "Create Project",
		Description: "Create a project in a workspace for one or more teams from workspace_id, name and team_ids. Status defaults to the first backlog status; lead and members must be workspace members. Returns the project. Needs member.",
		Path:        "/create",
		Method:      http.MethodPost,
	}, handlers.Create, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-project",
		Summary:     "Update Project",
		Description: "Change a project's name, status, lead, members, teams, dates, priority, labels or initiatives. Omitted fields keep their value; list fields replace the whole list. Returns the project. Needs member.",
		Path:        "/update",
		Method:      http.MethodPut,
	}, handlers.Update, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-project",
		Summary:     "Delete Project",
		Description: "Delete a project with its milestones, updates, links and documents. Its issues stay and lose the project. Needs admin.",
		Path:        "/delete",
		Method:      http.MethodDelete,
	}, handlers.Delete, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "archive-project",
		Summary:     "Archive Project",
		Description: "Archive a project so it is hidden from lists unless include_archived is set. Returns the project. Needs member.",
		Path:        "/archive",
		Method:      http.MethodPost,
	}, handlers.Archive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "unarchive-project",
		Summary:     "Unarchive Project",
		Description: "Bring an archived project back into lists. Returns the project. Needs member.",
		Path:        "/unarchive",
		Method:      http.MethodPost,
	}, handlers.Unarchive, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "trash-project",
		Summary:     "Trash Project",
		Description: "Move a project to the trash; it is hidden from lists unless include_trashed is set. Returns the project. Needs member.",
		Path:        "/trash",
		Method:      http.MethodPost,
	}, handlers.Trash, oapi.MCP)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "restore-project",
		Summary:     "Restore Project",
		Description: "Take a project out of the trash. Returns the project. Needs member.",
		Path:        "/restore",
		Method:      http.MethodPost,
	}, handlers.Restore, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-project-statuses",
		Summary:     "List Project Statuses",
		Description: "List the project statuses of a workspace, ordered by type (backlog, planned, started, paused, completed, canceled) then position.",
		Path:        "/statuses/list",
		Method:      http.MethodGet,
	}, handlers.ListStatuses, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-project-status",
		Summary:     "Create Project Status",
		Description: "Add a project status to a workspace from workspace_id, name, type and color. Returns the status. Needs admin.",
		Path:        "/statuses/create",
		Method:      http.MethodPost,
	}, handlers.CreateStatus, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-project-status",
		Summary:     "Update Project Status",
		Description: "Rename, recolor, describe or reorder a project status. Its type cannot change. Needs admin.",
		Path:        "/statuses/update",
		Method:      http.MethodPut,
	}, handlers.UpdateStatus, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-project-status",
		Summary:     "Delete Project Status",
		Description: "Delete a project status after moving its projects to another status of the same workspace. Needs admin.",
		Path:        "/statuses/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteStatus, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-milestones",
		Summary:     "List Milestones",
		Description: "List a project's milestones in order, each with progress from the issues assigned to it.",
		Path:        "/milestones/list",
		Method:      http.MethodGet,
	}, handlers.ListMilestones, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-milestone",
		Summary:     "Create Milestone",
		Description: "Add a milestone to a project from project_id and name. Issues join it through update-issue with milestone_id. Returns the milestone. Needs member.",
		Path:        "/milestones/create",
		Method:      http.MethodPost,
	}, handlers.CreateMilestone, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-milestone",
		Summary:     "Update Milestone",
		Description: "Rename, describe, reschedule or reorder a milestone. Omitted fields keep their value. Returns the milestone. Needs member.",
		Path:        "/milestones/update",
		Method:      http.MethodPut,
	}, handlers.UpdateMilestone, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-milestone",
		Summary:     "Delete Milestone",
		Description: "Delete a milestone. Its issues stay in the project without a milestone. Needs member.",
		Path:        "/milestones/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteMilestone, oapi.MCP)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-project-updates",
		Summary:     "List Project Updates",
		Description: "List a project's status updates, newest first, with author, health and reactions.",
		Path:        "/updates/list",
		Method:      http.MethodGet,
	}, handlers.ListUpdates, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-project-update",
		Summary:     "Create Project Update",
		Description: "Post a status update on a project with a markdown body and a health. The project's health follows the latest update and its lead and members are notified. Returns the update. Needs member.",
		Path:        "/updates/create",
		Method:      http.MethodPost,
	}, handlers.CreateUpdate, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-project-update",
		Summary:     "Update Project Update",
		Description: "Edit the body or health of a project update. Only its author or an admin may.",
		Path:        "/updates/update",
		Method:      http.MethodPut,
	}, handlers.EditUpdate, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-project-update",
		Summary:     "Delete Project Update",
		Description: "Delete a project update. Only its author or an admin may. The project's health falls back to the previous update.",
		Path:        "/updates/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteUpdate, oapi.MCP)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-project-link",
		Summary:     "Create Project Link",
		Description: "Attach an http or https link with a label to a project. Returns the link. Needs member.",
		Path:        "/links/create",
		Method:      http.MethodPost,
	}, handlers.CreateLink, oapi.MCP)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-project-link",
		Summary:     "Update Project Link",
		Description: "Change a project link's URL, label or order. Returns the link. Needs member.",
		Path:        "/links/update",
		Method:      http.MethodPut,
	}, handlers.UpdateLink, oapi.MCP)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-project-link",
		Summary:     "Delete Project Link",
		Description: "Remove a link from a project. Needs member.",
		Path:        "/links/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteLink, oapi.MCP)
}
