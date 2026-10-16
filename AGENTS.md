# Zigzag

Zigzag is a self-hosted issue tracker modeled on Linear. It ships as one binary: a Go API that embeds and serves a React SPA, backed by PostgreSQL. It also ships an MCP server so agents can work issues through the same API.

The architecture, conventions and toolkits are copied from Unbind (https://github.com/unbindapp/unbind, MIT). Where this repo and Unbind differ, this repo's `docs/design/*.md` wins. Where this repo is silent, do what Unbind does. Do not invent new patterns, libraries or folder layouts when Unbind already has one for the same job.

## Repo layout

```
apps/web    React SPA: Vite, TanStack Router, TanStack Query, TanStack Form, Zod, shadcn (base-ui), Tailwind v4
apps/api    Go API: Huma v2 on chi, ent, goose migrations, MCP server. module github.com/unbindapp/zigzag-api
deploy/     docker compose for local Postgres
docs/design Design documents. Read them before touching code.
```

## Design documents (binding)

- `docs/design/architecture.md` request flow, auth, MCP, mail, config, dev environment
- `docs/design/data-model.md` every ent entity and field
- `docs/design/api.md` every operation: id, path, input, output, permission, MCP exposure
- `docs/design/ui.md` routes, layout primitives, spacing and typography rules, component inventory
- `docs/design/linear-import.md` the Linear export format and how it maps

If you need something these documents do not cover, follow Unbind's convention and add a line to the relevant document. Never change a documented decision silently.

## Hard rules

- Self-documenting code, near-zero comments. Names and structure carry meaning. A comment survives only for a trap the code cannot show, one terse line. Never narrate what code does, never put design rationale in code.
- Guard clauses. Bail early, happy path at the leftmost indent. No nested if/else pyramids. switch/case when dispatching on one value.
- Keep it simple. Reinvent the wheel, not the car: no library for a simple problem, a modern library for a complex common one.
- Security and reliability first. Every operation checks permissions in the service layer through the permissions repository. Handlers never check permissions themselves.
- Generated code is never edited by hand: `apps/api/ent/*` (ent), `*_iface.go` (ifacemaker), `apps/api/mocks/*` (mockery), `apps/web/src/lib/server/client.gen.ts` (generate-sdk), `apps/web/src/routeTree.gen.ts` (router plugin).
- Tests cover input and output behavior. No tests that assert a file exists or a function is defined.
- Every API operation is registered through `oapi.Register` with an `Action`, and either `oapi.MCP` or `oapi.NoMCP("reason")`. The MCP coverage test fails otherwise.
- An operation takes either query parameters or a JSON body, never both. No path parameters. Responses wrap in `{ "data": ... }`.
- The web talks to the API only through the generated client in `src/lib/server/client.gen.ts` and the query modules in `src/lib/queries/*`. No hand-written fetch calls.
- UI consistency is a correctness requirement. Use the layout primitives and spacing scale in `docs/design/ui.md`. If two screens show the same kind of thing, they use the same component with the same props. A one-off margin, font size or button size is a bug.
- Guard against the same mistake in both directions: if the API gains a field, the web client is regenerated and the UI shows it; if the UI needs a field, the API grows it.
- No Claude or AI attribution anywhere in code, commits or PRs.

## Workflow

- Run `make` targets, not ad hoc commands. `make help` lists them.
- After changing the API: `make gen-web-types` then `make api-test`, `make web-typecheck`, `make web-lint`.
- After changing ent schemas: `make api-ent`, then `make api-migrate NAME=<snake_case>` and review the SQL.
- Before finishing any task: `make check` (fmt, vet, test, typecheck, lint, type sync) must pass.
- Commit messages: `api: ...`, `web: ...`, `web | api: ...`, lowercase, imperative, short.

## MCP server

`apps/api/internal/mcpserver` serves `/mcp`. It has no tool code of its own: every tool is an API operation registered with `oapi.MCP`, called in process through the router, so a tool validates, authorizes and answers exactly like its endpoint. Clients authenticate with an OAuth grant (connected apps) or an API key. Both carry a role and an optional workspace scope. Tool name, title, description and input schema come from the operation's `OperationID`, `Summary`, `Description` and input struct. Write them for an agent to read, and use `doc` tags on every input field. If the way resources relate changes, update `instructions` in `mcpserver.go`.
