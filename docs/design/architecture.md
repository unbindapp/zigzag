# Architecture

Zigzag copies Unbind's shape. Read Unbind's `apps/api` and `apps/web` for any detail not covered here. A reference checkout lives at `/root/research/unbind` on the dev machine.

## One binary

- `apps/api/cmd/api` starts the HTTP server on `:8091` (`PORT`).
- The SPA build is copied to `apps/api/internal/web/dist` and embedded with `go:embed`. `internal/web/web.go` serves it as the chi `NotFound` fallback: hashed `assets/` are immutable, `config.json` is never cached, unknown paths fall back to `index.html` only for `Accept: text/html`.
- The SPA calls the API same-origin at `/api/go` (from `public/config.json`). The binary serves every API operation both at the root and under `/api/go` (the same router with the prefix stripped), so it works behind an ingress that strips the prefix, one that does not, and the Vite proxy. `/mcp`, `/oauth/*` and `/.well-known/*` are served at the root only; under `/api/go` they answer 404. The OpenAPI `servers` URL is `EXTERNAL_API_URL`.
- The root `Dockerfile` builds the SPA (`node:24-alpine`), embeds it and builds the binary (`golang:1.27-alpine`), and ships `/app/zigzag` on `distroless/static-debian12:nonroot` with `DATA_DIR=/data` as a volume. Time zone data is compiled in with `time/tzdata`.
- `/openapi.json` is the full spec. `/docs` renders `/openapi-public.json`, the operations an API key can call (`oapi.APIKeySpec`, also behind `cmd/openapi -api-key-only`).
- `apps/api/cmd/openapi` prints the OpenAPI spec offline. `scripts/gen-web-types.sh` feeds it to `apps/web/scripts/generate-sdk` to produce `src/lib/server/client.gen.ts` (zod schemas and a typed client). Both are committed; `scripts/check-web-types.sh` fails CI when stale.

## API layering (identical to Unbind)

```
internal/api/handlers/<group>   huma input/output structs, RegisterHandlers, calls one service method, maps errors with oapi.MapError
internal/services/<name>        business rules, permission checks, transactions
internal/repositories/<name>    ent queries and mutations; interfaces generated with ifacemaker; mocks with mockery
internal/repositories/repositories  facade exposing every repo, WithTx
internal/models                 response and shared input structs, Transform* functions from ent entities
internal/api/oapi               Register, Action, MCP/NoMCP, SessionOnly, Needs, MapError
internal/api/middleware         Authenticate, CSRF, RequestID, RealIP, Logger, Recoverer, RateLimiter, ErrorTransformer
internal/api/router             NewHumaConfig, RegisterRoutes (groups with prefix, tag, authed, sessionOnly)
internal/api/server             Server struct holding Cfg, Repository, services; AuthenticatedUser
internal/api/apictx             typed context keys, MCPCaller
internal/auth                   JWT access tokens (RSA key stored in DB), refresh tokens, cookies, CSRF, API keys, OIDC discovery
internal/oauthserver            OAuth 2.1 authorization server for connected apps (port of Unbind's: dynamic client registration, PKCE, CIMD, consent)
internal/mcpserver              MCP streamable HTTP at /mcp, tools derived from oapi.MCP operations
internal/infrastructure/database  ent client
internal/infrastructure/mail    SMTP sender (wneessen/go-mail) and templates
internal/infrastructure/storage  file storage for uploads (local directory under DATA_DIR)
internal/common/errdefs         ResponseError envelope, CustomError types, InternalError
internal/common/log             charmbracelet/log wrapper
internal/common/utils           slugs, random tokens, url joins
internal/common/markdown        plain text extraction, mention parsing
internal/scheduler              gocron jobs: cycle rollover, oauth cleanup, notification email digest, invitation expiry
config/                         caarlos0/env Config
ent/                            schema/, generated code, migrate/ (goose SQL, atlas.sum, migrate main)
```

Handlers use `self` as the receiver name, like Unbind. Handler packages are named `<group>_handler`, services `<name>_service`, repositories `<name>_repo`.

Error flow: services return `errdefs` errors (`ErrUnauthorized`, `NewCustomError(ErrTypeNotFound, ...)`, `NewCustomError(ErrTypeInvalidInput, ...)`, `ErrConflict`, `NewInternalError(cause, summary)`); handlers call `oapi.MapError`. Authorization failures on reads are masked as not found with `errdefs.MaskAsNotFound`.

## Authentication

Port Unbind's `internal/auth`, `internal/api/middleware/auth.go`, `handlers/auth`, `handlers/setup` with these changes:

- Token prefixes: API key `zz_`, OAuth access `zzat_`, OAuth refresh `zzrt_`.
- Session JWTs carry the user id as `sub` and a random `jti`; the issuer is the OAuth issuer (`EXTERNAL_UI_URL`), so `/.well-known/jwks.json` at the root verifies them. Deactivated users (`active=false`) are refused on every credential kind.
- Cookies: `__Host-access_token`, `__Host-refresh_token`, `__Host-csrf_token` when `COOKIE_SECURE`, plain names otherwise. Access token TTL 15 minutes, refresh 14 days, rotated transparently by the middleware.
- CSRF: double submit, `X-CSRF-Token` header on every cookie-authenticated mutation.
- `GET /setup/status` and `POST /setup/create-user` bootstrap the first user, who becomes a system admin and gets a first workspace named after the instance.
- Password hashing: bcrypt cost 12 (Unbind's `user_repo.Authenticate`).
- Password reset: `POST /auth/request-password-reset` emails a one-time link when mail is configured; `POST /auth/reset-password` consumes it. Tokens hashed in `password_reset_tokens`.
- Sessions, API keys and OAuth grants are the three credential kinds. API keys and OAuth grants are "narrowed" credentials: they carry a role and an optional workspace list and can never exceed the owner's own access. Session-only operations refuse them (`oapi.SessionOnly`).

## Authorization

Three workspace roles, each including the one before: `viewer` (read), `member` (create and edit issues, projects, comments, documents), `admin` (also delete, manage members, teams, labels, workflow states, settings). A user with `system_admin` holds `admin` on every workspace and may manage instance settings and users.

`permissions_repo.Check(ctx, userID, []PermissionCheck{{Role, WorkspaceID}})` is the one hard check. It returns nil when any check passes:

1. the user is a system admin, or
2. the user is a workspace member with a role that implies the required one,

and in both cases, when the context carries `APIKeyAccess`, the credential's role must imply the required role and the credential must be full access or list the workspace.

Team-scoped resources resolve to their workspace before the check. There is no per-team or per-project role.

`permissions_repo.GetUserPermissionSet` returns every workspace the user can see with the effective role, used by `whoami` and by list filtering (`WorkspaceIDsVisibleTo`).

Credential privileges (`KeyPrivilege`): none at launch. Keep the field and the plumbing (`oapi.Needs`, `HasPrivilege`) so later privileges slot in. API keys and OAuth grants store an empty list. While the list is empty, `KeyPrivilege` is a plain string in the OpenAPI schema (an empty enum would reject everything) and validation refuses any value.

## MCP

Port `internal/mcpserver` from Unbind unchanged in design: `buildTools` walks the OpenAPI, one `mcp.Server` per distinct offered tool set, `dispatcher` runs tool calls through the router with `apictx.WithMCPCaller`, verifier accepts OAuth access tokens and API keys. `coverage_test.go` fails when an operation an API key can reach has neither `oapi.MCP` nor `oapi.NoMCP`. Rate limit 600 requests/minute per caller. `/mcp` accepts request bodies up to 64 MiB so `import-linear` can take a real export.

The `instructions` text tells an agent: resources nest as workspace > team > issue; projects and labels belong to a workspace and are shared by its teams; identifiers look like `KEY-123` and most issue tools accept either `issue_id` or `identifier`; call `whoami` first; `list-issues` is filtered and paginated with an opaque cursor; archived and trashed issues are hidden unless asked for; creating an issue needs `team_id` and `title`, everything else defaults (state = team default, priority none); markdown is accepted in descriptions and comments; `@` mentions use `@[Name](user:<id>)`.

## Mail

`internal/infrastructure/mail` is Unbind's SMTP sender. Settings live in `system_settings.email_settings` (host, port, username, from address, from name, encryption `starttls|tls|none`). The password is stored in `system_settings.email_password_encrypted`, AES-256-GCM with the key from `ENCRYPTION_KEY` (32 bytes hex). Resend works as a plain SMTP relay: host `smtp.resend.com`, port 465 with `tls` or 587 with `starttls`, username `resend`, password the API key; the UI shows this as a preset under "Resend" that fills host, port, encryption and username. `email_service` renders templates (invitation, test, password reset, notification) with a text and an HTML part and sends them. Sending never blocks a request: callers enqueue through `email_service.SendAsync` which runs in a goroutine with a 20s timeout and logs failures.

## Uploads

`POST /uploads/create` accepts multipart (`workspace_id`, `file`) and stores it under `DATA_DIR/uploads/<workspace_id>/<uuid><ext>`, recording an `uploads` row; png, jpeg, gif and webp get their pixel size from the file header. `GET /uploads/get?upload_id=` streams it back after a workspace viewer check. Only safe types (raster images, pdf, plain text, common audio and video) are served `inline` with their content type; everything else, including html and svg, is served as an `attachment` of `application/octet-stream`. Every response carries `X-Content-Type-Options: nosniff` and `Content-Security-Policy: default-src 'none'; sandbox`, since uploads share the app origin. Max size `MAX_UPLOAD_BYTES` (default 25 MiB). Both are `NoMCP`. Images in markdown reference `/api/go/uploads/get?upload_id=<id>`.

## Config (`config/config.go`, caarlos0/env)

```
PORT                     default 8091
EXTERNAL_UI_URL          default http://localhost:5174, also the OAuth issuer and MCP resource origin
EXTERNAL_API_URL         default <EXTERNAL_UI_URL>/api/go when EXTERNAL_UI_URL is set, else http://localhost:8091
DATABASE_URL             optional postgres:// URL; when set it replaces the POSTGRES_* values
POSTGRES_HOST/PORT/USER/PASSWORD/DB/SSL_MODE   defaults localhost 54329 postgres postgres zigzag disable
COOKIE_SECURE            default true (false in dev)
TOKEN_AUDIENCE           default zigzag-api
ENCRYPTION_KEY           required, 64 hex chars
DATA_DIR                 default ./.data
MAX_UPLOAD_BYTES         default 26214400
INJECT_DEV_ORIGINS       default false
LOG_LEVEL                default info
```

No Redis. Rate limiting (`middleware.RateLimiter`, same `Limit(name, n, window)` API as Unbind) and the client metadata document cache are in-process (the single binary runs as one replica). `login`, `request-password-reset` and `accept-invitation` are limited to 10 requests per minute per client address (`middleware.RateLimitOperations`, applied to every group and keyed by operation id).

## Scheduler (`internal/scheduler`, gocron)

- every 10 minutes: close cycles whose `ends_at` passed, create the next cycle for teams with cycles enabled, move unfinished issues to the next cycle when `cycle_auto_move` is on
- every hour: oauth cleanup (expired codes, tokens), expired invitations and password reset tokens
- every 5 minutes: email notifications not yet sent, batched per user (one email per user per batch). A notification is emailed when it is still unread, not archived, not snoozed, older than 2 minutes and younger than 24 hours, and the user is active with `email_enabled`; the batch marks `emailed_at` on what it sent and on what it skipped for settings, and leaves the rest pending when SMTP fails. Nothing is read while mail is not configured.
- every 24 hours: `issue_due_soon` to the assignee of open, unarchived issues due today or tomorrow (UTC dates within the next 24 hours), at most once per issue and user in 48 hours
- every hour: hard delete issues trashed more than 30 days ago (`issue_service.PurgeTrashed`)
- every hour: auto close (`issue_service.AutoCloseStaleIssues`) moves triage, backlog and unstarted issues not updated for the team's `auto_close_months` to its Canceled state (else its first canceled state) and sets `auto_closed_at`; auto archive (`issue_service.AutoArchiveDoneIssues`) archives completed and canceled issues closed more than `auto_archive_months` ago and sets `auto_archived_at`; both record system history
- jobs run with `permissions_repo.WithSystemCaller` and in singleton mode, so a slow run is never overlapped

## Dev environment

- `deploy/compose/docker-compose.yaml`: Postgres 16 on `127.0.0.1:54329`, container `zigzag_postgres`, db `zigzag`, compose project `zigzag` (pinned so it never collides with another `deploy/compose` project on the same machine).
- `make dev-infra` starts it. `make dev-api` runs the API (`COOKIE_SECURE=false INJECT_DEV_ORIGINS=true`). `make dev-web` runs Vite on `:5174` proxying to `:8091`.
- `make app` builds the SPA, copies it into the embed dir and builds `apps/api/bin/zigzag`.
- `make check` runs: `go vet`, `gofmt -l`, `go test ./...`, `check-web-types`, `npm run typecheck`, `npm run lint`, `npm run test`.
- Migrations: `make api-migrate NAME=x` uses embedded-postgres like Unbind, through `apps/api/scripts/migrate.sh`. `initdb` refuses root, so as root the script diffs against a scratch database `zigzag_migrate_<random>` it creates on the running `zigzag_postgres` container and drops afterwards. `make api-migrate-checksum` regenerates `atlas.sum`. Migrations run with goose on startup. The first, hand-written migration `..._extensions.sql` creates `pg_trgm`; atlas applies migrations to its dev database while diffing, so the extension must exist before the generated migration that uses `gin_trgm_ops`.
- Repository tests run on in-memory SQLite. With `ZIGZAG_TEST_DATABASE_URL=postgres://postgres:postgres@127.0.0.1:54329/postgres?sslmode=disable` the issue repository suite also runs on Postgres (search, row locks, keyset paging) in a scratch database `zigzag_test_<random>` it migrates and drops.
- First run: open the UI, `/welcome` creates the first user.
