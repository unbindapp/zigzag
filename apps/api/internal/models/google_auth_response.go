package models

type GoogleAuthSettingsResponse struct {
	Enabled      bool   `json:"enabled" doc:"Whether Sign in with Google is offered"`
	ClientID     string `json:"client_id" doc:"OAuth client id from Google Cloud"`
	HostedDomain string `json:"hosted_domain" doc:"Only Google Workspace accounts of this domain may sign in; empty allows any Google account"`
	Configured   bool   `json:"configured" doc:"True when a client id and secret are stored"`
	SecretPrefix string `json:"secret_prefix" doc:"First characters of the client secret, empty without one. Never the full secret."`
	RedirectURI  string `json:"redirect_uri" doc:"Authorized redirect URI to register on the Google OAuth client"`
}

type AuthProviderStatus struct {
	Enabled bool `json:"enabled"`
}

type AuthProvidersResponse struct {
	Google AuthProviderStatus `json:"google"`
}
