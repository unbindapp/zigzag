package integrations_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zigzagapp-org/zigzag-api/internal/api/oapi"
	"github.com/zigzagapp-org/zigzag-api/internal/api/server"
)

const maxWebhookBodyBytes = 1 << 20

type HandlerGroup struct {
	srv *server.Server
}

func RegisterHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-github-integration",
		Summary:     "Get GitHub Integration",
		Description: "Show whether GitHub is connected to a workspace, the webhook URL to add on GitHub, the secret's prefix and the repositories that delivered events. Needs admin.",
		Path:        "/github/get",
		Method:      http.MethodGet,
	}, handlers.GetGithub, oapi.SessionOnly)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "create-github-integration",
		Summary:     "Create GitHub Integration",
		Description: "Connect GitHub to a workspace by generating a webhook secret. The secret is returned once. Needs admin.",
		Path:        "/github/create",
		Method:      http.MethodPost,
	}, handlers.CreateGithub, oapi.SessionOnly, oapi.Confirm)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "regenerate-github-integration-secret",
		Summary:     "Regenerate GitHub Webhook Secret",
		Description: "Replace the workspace's GitHub webhook secret. Deliveries signed with the old secret are refused from now on. The new secret is returned once. Needs admin.",
		Path:        "/github/regenerate-secret",
		Method:      http.MethodPost,
	}, handlers.RegenerateGithubSecret, oapi.SessionOnly, oapi.Confirm)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "delete-github-integration",
		Summary:     "Delete GitHub Integration",
		Description: "Disconnect GitHub from a workspace. Deliveries are refused and the repository list is cleared; attachments already on issues stay. Needs admin.",
		Path:        "/github/delete",
		Method:      http.MethodDelete,
	}, handlers.DeleteGithub, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-github-integration",
		Summary:     "Update GitHub Integration",
		Description: "Store or clear the GitHub token Issues sync uses to send changes made in Zigzag back to GitHub and to import existing issues. A fine-grained token with Issues read and write on the synced repositories. Only its prefix is shown afterwards. Needs admin.",
		Path:        "/github/update",
		Method:      http.MethodPut,
	}, handlers.UpdateGithub, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "list-github-repo-syncs",
		Summary:     "List Synced GitHub Repositories",
		Description: "List the GitHub repositories whose issues sync both ways with a workspace, with their team, project, linked issue count, import status and last error. Needs admin.",
		Path:        "/github/repo-syncs/list",
		Method:      http.MethodGet,
	}, handlers.ListGithubRepoSyncs, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-github-repo-sync",
		Summary:     "Get Synced GitHub Repository",
		Description: "Show one synced GitHub repository, including the status and counts of its running or last import. Needs admin.",
		Path:        "/github/repo-syncs/get",
		Method:      http.MethodGet,
	}, handlers.GetGithubRepoSync, oapi.SessionOnly)

	oapi.Register(grp, oapi.Create, huma.Operation{
		OperationID: "add-github-repo-sync",
		Summary:     "Add Synced GitHub Repository",
		Description: "Start syncing a GitHub repository's issues with a team. Zigzag creates or reuses a project named owner/name for them. New issues, edits, labels, assignees, state and comments from the webhook flow in; changes in Zigzag go back when a token is stored. Needs admin.",
		Path:        "/github/repo-syncs/add",
		Method:      http.MethodPost,
	}, handlers.AddGithubRepoSync, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "backfill-github-repo-sync",
		Summary:     "Import Existing GitHub Issues",
		Description: "Import the repository's existing open issues, or all issues with sync_closed, and their comments in the background. Issues already linked are updated. Poll get-github-repo-sync for the status. Needs admin and a stored token.",
		Path:        "/github/repo-syncs/backfill",
		Method:      http.MethodPost,
	}, handlers.BackfillGithubRepoSync, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "remove-github-repo-sync",
		Summary:     "Remove Synced GitHub Repository",
		Description: "Stop syncing a GitHub repository. Its issues stay in Zigzag without the link to GitHub. Needs admin.",
		Path:        "/github/repo-syncs/remove",
		Method:      http.MethodDelete,
	}, handlers.RemoveGithubRepoSync, oapi.SessionOnly)
}

func RegisterPublicHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID:  "github-webhook",
		Summary:      "GitHub Webhook",
		Description:  "Receive a GitHub webhook delivery for a workspace. Authenticated by the X-Hub-Signature-256 header, an HMAC-SHA256 of the raw body keyed with the workspace's GitHub secret. Handles ping, pull_request and push, and issues and issue_comment for synced repositories; other events are acknowledged and ignored.",
		Path:         "/github/webhook",
		Method:       http.MethodPost,
		MaxBodyBytes: maxWebhookBodyBytes,
	}, handlers.GithubWebhook, oapi.Public, oapi.NoMCP("GitHub calls it"))
}
