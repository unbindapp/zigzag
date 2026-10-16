# Linear import

`POST /imports/linear` takes a Linear export in the JSON format below and imports it into one workspace. The same format is what `scripts/linear-export` (a Node script that reads from the Linear MCP server's output files) and the dev machine's MCP-driven export produce.

## Request

```
{
  "workspace_id": uuid,
  "dry_run": bool,                default true
  "restore_archived": bool,       import archived issues and projects as active (archived_at cleared). default false
  "restore_trashed": bool,        import trashed issues and projects as active. default false
  "create_missing_users": bool,   create placeholder accounts for users not in the workspace. default true
  "export": LinearExport
}
```

Response `data`: an `ImportReport`:

```
{
  "dry_run": bool,
  "import_id": uuid,
  "counts": { "<entity>": { "created": n, "updated": n, "skipped": n } },   entities: users, teams, workflow_states, labels, projects, milestones, cycles, issues, comments, attachments, relations, history
  "warnings": [ { "entity": "issue", "external_id": "UNB-12", "message": "..." } ],
  "user_mapping": [ { "linear_id": "...", "email": "...", "user_id": uuid|null, "action": "matched|created|skipped" } ]
}
```

A dry run runs the whole import inside a transaction and rolls it back, so the report is exact. A real import is one transaction per entity group (users, teams and states, labels, projects, cycles, issues, comments and relations) and is idempotent: rows carry `import_source="linear"` and `import_external_id=<linear uuid>`, and a second run updates instead of duplicating. Issues keep their Linear numbers; the team's `issue_counter` is raised to the highest imported number.

## Mapping

| Linear | Zigzag |
| --- | --- |
| workspace | the target workspace (never created) |
| team (key, name, description, icon, color) | team with the same key; created if missing |
| workflow state (name, type, color, position) | team workflow state of the same type and name; created if missing; team `default_state_id` = first `unstarted` state when the team is new. Linear's `duplicate` type maps to `canceled` |
| label (name, color, parent, isGroup, team) | workspace label when `teamId` is null, team label otherwise; groups become `is_group` with `group_single_select` from `groupType == "singleSelect"` |
| project (name, summary, description, status, lead, priority, dates, sortOrder, teams, members, labels, archived, trashed) | project; status matched by type then name; `trashed` becomes `trashed_at = updatedAt` unless `restore_trashed` |
| milestone | project milestone |
| cycle (number, name, startsAt, endsAt, completedAt) | team cycle with the same number |
| issue | issue with the same team and `number`; `priority` value copied; `estimate`; `state` by state id or by `(name, type)`; labels by name; `parentId` resolved after all issues exist; `archivedAt` kept unless `restore_archived`; `url` stored in `import_url`; `gitBranchName` recomputed |
| issue.relations | `blocks`, `relatedTo`, `duplicateOf` |
| issue.stateHistory | one `issue_history` row per transition with `source=import`, actor nil |
| issue.attachments | attachment with `source_type=import` and the original url |
| comment (body, author, createdAt, parentId) | comment; author by user mapping; `createdAt` preserved |
| user (email, name, displayName, isAdmin, isActive) | matched by email (case insensitive) to an existing user, else created inactive with no password when `create_missing_users`; added to the workspace as `member` (`admin` when `isAdmin`) and to each team they appear on; the Linear bot user (`*@linear.linear.app`) is skipped and its authored content keeps `author_id = nil` |

Unknown fields are ignored. Dates are RFC 3339.

## Export format

```
LinearExport {
  "exported_at": string,
  "workspace": { "id", "name", "url" },
  "users": [ { "id", "name", "email", "displayName", "avatarUrl", "isAdmin", "isGuest", "isActive" } ],
  "teams": [ { "id", "key", "name", "description", "icon", "color", "createdAt", "cyclesEnabled", "timezone" } ],
  "workflow_states": [ { "id", "teamId", "name", "type", "color", "position" } ],
  "labels": [ { "id", "name", "color", "description", "parentId", "isGroup", "groupType", "teamId", "archivedAt" } ],
  "projects": [ { "id", "identifier", "name", "summary", "description", "url", "status": { "id", "name", "type" }, "leadId", "priority", "startDate", "targetDate", "sortOrder", "createdAt", "updatedAt", "startedAt", "completedAt", "canceledAt", "archivedAt", "trashed", "teamIds", "memberIds", "labelIds" } ],
  "milestones": [ { "id", "projectId", "name", "description", "targetDate", "sortOrder" } ],
  "cycles": [ { "id", "teamId", "number", "name", "description", "startsAt", "endsAt", "completedAt" } ],
  "issues": [ {
    "id", "identifier", "number", "teamId", "title", "description", "priority", "estimate", "url", "gitBranchName",
    "createdAt", "updatedAt", "archivedAt", "completedAt", "startedAt", "canceledAt", "triagedAt", "dueDate",
    "stateId", "stateName", "stateType", "labelIds", "labelNames", "creatorId", "assigneeId", "projectId", "milestoneId", "parentId", "cycleId", "sortOrder",
    "attachments": [ { "id", "title", "subtitle", "url", "createdAt" } ],
    "relations": { "blocks": [issue ids], "blockedBy": [issue ids], "relatedTo": [issue ids], "duplicateOf": issue id | null },
    "stateHistory": [ { "stateId", "stateName", "stateType", "startedAt", "endedAt" } ],
    "comments": [ { "id", "body", "authorId", "authorName", "createdAt", "updatedAt", "parentId", "quotedText" } ]
  } ]
}
```

Issue ids in `parentId`, `relations` and `comments.parentId` are Linear uuids. `priority` is the integer 0..4.

## Known limits of the MCP source

The Linear MCP server lists archived issues (`includeArchived`) but does not list trashed issues; `get_issue` on a trashed identifier answers "Could not find referenced Issue". An export built from MCP therefore cannot contain trashed issues; `restore_trashed` only matters for exports produced from Linear's own CSV or GraphQL exports. The importer accepts both.
