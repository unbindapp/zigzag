package system_handler

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
		OperationID: "get-system-settings",
		Summary:     "Get System Settings",
		Description: "Get the instance settings and whether email is configured. System admins only.",
		Path:        "/settings",
		Method:      http.MethodGet,
	}, handlers.GetSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-system-settings",
		Summary:     "Update System Settings",
		Description: "Change the instance name or public signup. System admins only.",
		Path:        "/settings",
		Method:      http.MethodPut,
	}, handlers.UpdateSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-email-settings",
		Summary:     "Get Email Settings",
		Description: "Get the SMTP relay settings. The password is never returned. System admins only.",
		Path:        "/email",
		Method:      http.MethodGet,
	}, handlers.GetEmailSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-email-settings",
		Summary:     "Update Email Settings",
		Description: "Set the SMTP relay. For Resend use host smtp.resend.com, port 465 with tls or 587 with starttls, username resend and the API key as password. System admins only.",
		Path:        "/email",
		Method:      http.MethodPut,
	}, handlers.UpdateEmailSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "clear-email-settings",
		Summary:     "Clear Email Settings",
		Description: "Remove the SMTP relay and its password. Mail stops being sent. System admins only.",
		Path:        "/email",
		Method:      http.MethodDelete,
	}, handlers.ClearEmailSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Invoke, huma.Operation{
		OperationID: "send-test-email",
		Summary:     "Send Test Email",
		Description: "Send a test message to your own address through the configured relay. System admins only.",
		Path:        "/email/test",
		Method:      http.MethodPost,
	}, handlers.SendTestEmail, oapi.SessionOnly)

	oapi.Register(grp, oapi.Read, huma.Operation{
		OperationID: "get-google-auth-settings",
		Summary:     "Get Google Auth Settings",
		Description: "Get the Sign in with Google settings and the redirect URI to register in Google Cloud. The client secret is never returned. System admins only.",
		Path:        "/google-auth",
		Method:      http.MethodGet,
	}, handlers.GetGoogleAuthSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Update, huma.Operation{
		OperationID: "update-google-auth-settings",
		Summary:     "Update Google Auth Settings",
		Description: "Set the Google OAuth client and turn Sign in with Google on or off. System admins only.",
		Path:        "/google-auth",
		Method:      http.MethodPut,
	}, handlers.UpdateGoogleAuthSettings, oapi.SessionOnly)

	oapi.Register(grp, oapi.Delete, huma.Operation{
		OperationID: "clear-google-auth-settings",
		Summary:     "Clear Google Auth Settings",
		Description: "Remove the Google OAuth client and its secret. Sign in with Google stops working; linked accounts keep their link. System admins only.",
		Path:        "/google-auth",
		Method:      http.MethodDelete,
	}, handlers.ClearGoogleAuthSettings, oapi.SessionOnly)
}
