<div align="center">

<img src="apps/web/public/favicon-512.png" alt="Zigzag" width="96" height="96" />

# Zigzag

**Self-hosted issue tracking, with an MCP server built in.**

Workspaces, teams, issues, projects, cycles, labels, views, documents, inbox. One binary, one Postgres.

</div>

---

> [!IMPORTANT]
> **In memoriam: Zigzag, 10/16/26 - 10/16/26**
>
> Zigzag has been acquired by [Ricrac](https://github.com/unbindapp/ricrac). This repository is archived and will not receive updates.
>
> To all of our faithful users: thank you. Please migrate to [Ricrac](https://github.com/unbindapp/ricrac).

---

## What you get

- Issues with sub-issues, relations, labels, priorities, estimates, due dates, comments, reactions, attachments and full history
- Projects with milestones, status updates and health, initiatives, documents
- Cycles with automatic rollover and burn-up
- Workflow states, label groups, templates, saved views, favorites
- Inbox notifications and email digests over SMTP (works with Resend)
- Workspaces with admin, member and viewer roles, invitations, API keys, connected apps
- Sign in with Google, set up by a system admin in the UI
- An MCP server at `/mcp` where every API operation is a tool, scoped by the same roles
- Import from Linear, archived issues included, idempotent and dry-runnable
- GitHub webhooks: pull requests, branches and commits that mention an issue link to it and move it from started to done; a commit that fixes an issue closes it when it lands on the default branch
- GitHub Issues sync: issues, comments, labels, assignees and state of chosen repositories sync both ways, with an import of existing issues

## Run it

```sh
docker run -d --name zigzag \
  -p 8091:8091 \
  -e DATABASE_URL=postgres://user:pass@host:5432/zigzag \
  -e EXTERNAL_UI_URL=https://issues.example.com \
  -e ENCRYPTION_KEY=$(openssl rand -hex 32) \
  ghcr.io/zigzagapp-org/zigzag
```

Open the URL. The first visit creates the first user and workspace.

| Variable             | Purpose                                                                                   |
| -------------------- | ----------------------------------------------------------------------------------------- |
| `DATABASE_URL`       | Postgres connection string                                                                |
| `EXTERNAL_UI_URL`    | Public URL of the instance, used in links and emails                                      |
| `ENCRYPTION_KEY`     | 64 hex chars, encrypts the SMTP password, Google client secret and GitHub secrets at rest |
| `PORT`               | Listen port, default `8091`                                                               |
| `GITHUB_API_URL`     | GitHub REST API base for Issues sync, default `https://api.github.com`                    |
| `MAX_UPLOAD_BYTES`   | Largest accepted upload, default `26214400` (25 MiB)                                      |
| `UPLOADS_IMPORT_DIR` | One-time migration: an old `DATA_DIR/uploads` directory to copy into Postgres on startup  |

### Uploads

Uploads are stored in Postgres, so the container needs no volume. Instances that kept uploads on disk under `DATA_DIR/uploads` can move them once: mount the old volume (at `/data`, say) and set `UPLOADS_IMPORT_DIR=/data/uploads` for one boot, or run `go run ./cmd/cli uploads:import-dir <dir>` from `apps/api` against the same database. Either one stores the file of every upload that has no bytes in the database yet, skips the rest and logs how many it imported, found missing and failed. Running it again is safe. Remove the variable and the volume afterwards.

### Google sign-in

A system admin turns it on under Instance settings, Authentication. Create an OAuth client of type Web application in Google Cloud with the redirect URI `<EXTERNAL_UI_URL>/api/go/auth/google/callback` (the page shows it with a copy button), then paste the client ID and secret. An optional hosted domain limits sign-in to one Google Workspace domain. Google accounts sign in to the account with the same verified email; someone without an account gets one only with a pending invitation, which is accepted on the way in, or when public signup is on. A user without an avatar gets their Google profile photo on sign-in, once they belong to a workspace.

## Connect an agent

Create an API key under your account settings, then point any MCP client at `https://<host>/mcp` with `Authorization: Bearer <key>`. Connected apps can also use OAuth with dynamic client registration. Tools are limited to the key's role and workspaces; a viewer key is offered read tools only.

## Develop

```sh
corepack enable  # pnpm, pinned in apps/web/package.json
make web-install # pnpm install in apps/web
make dev-infra   # postgres in docker
make dev-api     # api on :8091
make dev-web     # vite on :5174
make check       # fmt, vet, lint, tests, type sync
```

Go API with Huma and ent in `apps/api`, React SPA with TanStack Router and Query in `apps/web`. The SPA is embedded in the Go binary. See `AGENTS.md` for conventions.

## License

MIT. See [LICENSE](LICENSE).
