## What is this?

Zigzag is an open-source, self-hosted issue tracker modeled on Linear. This is the monorepo for it. It ships as one binary: a Go API that embeds and serves a React SPA, backed by PostgreSQL, with an MCP server so agents can work issues through the same API.

The architecture, conventions and toolkits are copied from Unbind (https://github.com/unbindapp/unbind). Where this repo and Unbind differ, this file wins. Where this repo is silent, do what Unbind does.

## Repo Structure:

There are 2 main apps in the repo:

### 1- apps/web:

The UI. A React app using Vite. Tanstack Router, Tanstack Query, Tanstack Forms, Tanstack Virtual, Zod, shadcn (on base-ui), Tailwind v4, TipTap for markdown editing and dnd-kit for drag and drop are the core libraries. The UI is served by the API server. The API client in `src/lib/server/client.gen.ts` is generated from the API's OpenAPI spec with `make gen-web-types`; every call goes through it and the query modules in `src/lib/queries`.

### 2- apps/api

The server built with Go. Huma on chi, ent for the database, goose for migrations. The UI interacts with this server to manage workspaces, teams, issues, projects, cycles, labels, notifications and everything else. It also serves the static UI and the MCP server. Module `github.com/zigzagapp-org/zigzag-api`.

### Conventions that are not in the code

- API: operations are `oapi.Register(grp, action, huma.Operation{...}, handler, opts...)` with paths `<group>/<verb>` (`/list`, `/get`, `/create`, `/update`, `/delete`, named actions). Reads are `GET` with query parameters, mutations are `POST`/`PUT`/`DELETE` with a JSON body, never both. The one exception is `github-webhook`: GitHub posts its own raw payload, so the workspace comes in the query string. Responses wrap in `{ "data": ... }`, paginated lists add `next_cursor`. Pointer scalars in responses carry `nullable:"true"`. Update inputs are never nullable: the nil uuid clears an id field, an empty string clears a date, `estimate: -1` clears the estimate. Response structs are never recursive. Permissions are checked only in services via `repo.Permissions().Check`; forbidden reads are masked as not found.
- Roles: `viewer` reads, `member` creates and edits, `admin` also deletes and manages people and settings. `system_admin` users hold admin everywhere. API keys and OAuth grants narrow a user to a role and a list of workspaces.
- Web: page structure comes only from `components/layout` (`PageContainer`, `PageHeader`, `PageToolbar`, `SettingsPage`, `SettingsSection`, `EmptyState`, `ErrorCard`, `ListSkeleton`, `SplitView`). Spacing: `gap-1.5` icon to label, `gap-2` controls in a row, `gap-3` cards, `gap-4` form fields, `gap-6` sections, `p-4` cards, rows `h-9`. Type: page title `text-xl font-semibold` via `PageHeader`, section title `text-base font-semibold`, body `text-sm`, secondary `text-sm text-muted-foreground`, captions and identifiers `text-xs`, identifiers `font-mono`. Buttons only from `components/ui/button`, sizes `sm|default|icon`. Icons lucide `size-4` (`size-3.5` in `text-xs`, `size-5` in headers). No arbitrary Tailwind values, no raw `<button>`/`<input>`. Every list has skeleton, error and empty states; every mutation failure toasts; icon-only buttons have `aria-label`. `npm run lint` enforces the mechanical parts.
- GitHub: one webhook secret per workspace (`/integrations/github/*`, session only), stored encrypted because verifying `X-Hub-Signature-256` needs the secret itself. The public `github-webhook` links pull requests and branches to the issues their branch, title or body mention (`internal/integrations/github/identifiers.go`) as `github` and `github_branch` attachments, and moves issues to the team's `github_pr_opened_state_id` / `github_pr_merged_state_id` (default first started / first completed state). Changes are recorded as `source=system` history with a `github` field change.
- GitHub Issues sync: a workspace stores an encrypted GitHub token (`update-github-integration`) and adds repositories (`GithubRepoSync`, `full_name` unique per instance) bound to a team and a project named `owner/name`. `issues` and `issue_comment` deliveries of synced repositories create and update issues and comments (`GithubIssueLink`, `GithubCommentLink`, idempotent on node and comment ids); `users.github_login` maps GitHub users to Zigzag users, unmatched authors stay nil and comments get a `**@login on GitHub**` first line. Inbound work runs with `github.WithInbound` on the context, which adds the `github` history change and stops `github_sync_service` from echoing it back. Changes made in Zigzag reach GitHub through `internal/services/githubsync` after the issue or comment service commits, asynchronously through `internal/integrations/github/client.go`; failures land in `GithubRepoSync.last_error`. Backfill (`backfill-github-repo-sync`) pages through the repository's issues in a goroutine and reports counts on the row.
- Linear import: `POST /imports/linear` takes the export format in `apps/api/internal/importers/linear` (see its types and `testdata`), runs as a dry run by default, is idempotent on `import_external_id`, keeps issue numbers and can restore archived issues.
- Uploads: `internal/infrastructure/storage` stores them under `DATA_DIR/uploads` or, when `S3_BUCKET` is set, in an S3 bucket under the same `uploads/<workspace_id>/<uuid><ext>` keys behind an optional `S3_KEY_PREFIX` (applied only inside the S3 store; database rows keep the unprefixed key). `get-upload` always streams through the API, never a presigned redirect. With S3 on, `UploadService.MigrateFrom` copies uploads missing from the bucket out of `DATA_DIR/uploads` on every boot.

## MCP Server

Zigzag ships an MCP server inside `apps/api`, served at `/mcp` (code in `apps/api/internal/mcpserver`). It has no tool code of its own: every tool is an API operation, called in process through the API router, so a tool validates, authorizes and answers exactly like its endpoint. Clients authenticate with an OAuth grant (Connected Apps) or an API key, and both are limited to the role (viewer, member, admin) and the workspaces the user picked.

Keep it in sync with the API:

- Every operation an API key can reach must be registered with `oapi.MCP` (it becomes a tool) or `oapi.NoMCP("reason")`. `go test ./internal/mcpserver/` fails until a new endpoint picks one. Public and session only operations can never be tools.
- The tool name, title, description, input schema and hints come from the operation's `OperationID`, `Summary`, `Description`, input struct and `oapi` action. Write them for an agent to read, and use `doc` tags on input fields.
- A tool takes either query parameters or a JSON body, never both. Path, header and cookie parameters are not supported.
- Response structs are never recursive; the web client generator cannot type them.
- Opt out anything that returns or accepts credentials, streams, or only makes sense in the UI.
- If the way resources relate changes, update the `instructions` text in `mcpserver.go`.

In the dev environment the server is at `http://localhost:8091/mcp`. Test it with an API key as the bearer token.

## General Rules:

- Keep it simple. Do not overcomplicate things.
- Follow each app's own conventions. `apps/web` has a design system and a certain way of doing things, follow it as much as possible. The spacing and type scale below is not a suggestion: if two screens show the same kind of thing, they use the same component with the same props.
- Security and reliability are extremely important. Multiple workspaces with different members share one instance, do not treat it like a hobby project. Every operation checks permissions in the service layer through the permissions repository, never in handlers.
- `apps/web` and `apps/api` are tightly coupled. When you are changing one, be mindful of the consequences for the other. After changing the API, regenerate the web client.
- Do not start editing code in response to a question. We'll tell you when to edit code.
- Do not leave paragraphs of comments on top of the code. Avoid comments as much as possible with understandable function names and code. If they are necessary even then, make them concise. Remove such comments when you come by them in the codebase. Comments always move with code, not be left behind.
- Use guard statement patterns in any code you write.
- Do not write any useless tests, tests should cover input output validation - not useless things like "main.go contains func main()".
- Do not edit generated code directly: mocks generated with mockery, interfaces generated with ifacemaker, the ent folder generated by ent, `client.gen.ts` and `routeTree.gen.ts` in the web app. Not edited by hand.
- Reinvent the wheel but do not reinvent the car. If you are solving a simple problem do not introduce a library. If you are solving a complex but a common problem, there is likely a modern library for it, if so, use it.
- If we are missing a glaring issue when we ask you to do something, do not hesitate to point it out.
- Never commit or push code unless explicitly asked to do so.
- Never make a PR unless explicitly asked to do so.
- Do not insert yourself into our code, commits or PRs in any way. Our codebase is not your ad space.
- After you make code changes, run the pre-commit hooks (`prek run --all-files`, or `pre-commit run --all-files`) and `make check`, and fix anything they report. CI runs the same checks.
- Keep this file and the README in sync with the code.

## Commit Messages

Commit subjects are lowercase conventional commits, short and imperative:

    feat: add issue relations
    fix: keep imported descriptions on first open
    chore: bump ent

Add a body only when the subject cannot carry it, and keep it plain. No em dashes, no filler.

## Local Dev Environment

Postgres runs in docker compose (`deploy/compose/docker-compose.yaml`, container `zigzag_postgres`, `127.0.0.1:54329`, database `zigzag`):

    make dev-infra      # start postgres
    make dev-api        # API on :8091
    make dev-web        # Vite on :5174, proxying /api/go to the API
    make app            # single binary with the SPA embedded, apps/api/bin/zigzag
    make check          # fmt, vet, tests, type sync, typecheck, lint

The API needs `ENCRYPTION_KEY` (64 hex chars) and, for local http, `COOKIE_SECURE=false` and `INJECT_DEV_ORIGINS=true`. Open http://localhost:5174; the first visit creates the first user and workspace. API docs are at http://localhost:8091/docs.

Migrations: `make api-migrate NAME=what_changed` after a schema change, then review the SQL and run `make api-migrate-checksum`. Migrations apply on startup.

### Browser demos and UI verification

Use a headless browser on this machine. Drive Chromium with Playwright from a directory outside the repo (for example `~/demos/<task>-demo/`, `npm init -y && npm i playwright-core`, `npx playwright install chromium`), launch headless with `--no-sandbox`, viewport 1440x900, sign in, do the flow like a user would, take screenshots at each step and hand back their paths. Never install Playwright into `apps/web/node_modules`.
