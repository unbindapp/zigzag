package integrations_handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/internal/api/oapi"
	"github.com/unbindapp/zigzag-api/internal/api/server"
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
}

func RegisterPublicHandlers(server *server.Server, grp *huma.Group) {
	handlers := &HandlerGroup{srv: server}

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID:  "github-webhook",
		Summary:      "GitHub Webhook",
		Description:  "Receive a GitHub webhook delivery for a workspace. Authenticated by the X-Hub-Signature-256 header, an HMAC-SHA256 of the raw body keyed with the workspace's GitHub secret. Handles ping, pull_request and push; other events are acknowledged and ignored.",
		Path:         "/github/webhook",
		Method:       http.MethodPost,
		MaxBodyBytes: maxWebhookBodyBytes,
	}, handlers.GithubWebhook, oapi.Public, oapi.NoMCP("GitHub calls it"))
}
