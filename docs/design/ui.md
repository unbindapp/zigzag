# UI

The web app is Unbind's `apps/web` toolkit with Linear's information architecture. Copy Unbind's `components/ui/*`, `lib/server/*`, `lib/config.ts`, `lib/hooks/*`, `components/providers/*`, `globals.css` structure, `scripts/generate-sdk`, `vite.config.ts`, `eslint.config.mjs`, `tsconfig.json`, `components.json` and `index.html` shell. Remove what Zigzag does not need (staged changes, templates, logs, metrics, terminal, k8s things). Keep the base-ui shadcn style, DM Sans and Geist Mono fonts, `cn`, `cva`, `next-themes` theme switching, `motion`, `cmdk`, `@tanstack/react-hotkeys`, `@tanstack/react-virtual`, `zustand`, `usehooks-ts`, `date-fns`, `fuzzysort`, `lucide-react`.

Additional libraries allowed, nothing else without updating this document:
- `@dnd-kit/core`, `@dnd-kit/sortable` for board and list drag and drop
- `@tiptap/react`, `@tiptap/starter-kit`, `@tiptap/extension-link`, `@tiptap/extension-task-list`, `@tiptap/extension-task-item`, `@tiptap/extension-image`, `@tiptap/extension-mention`, `@tiptap/extension-placeholder`, `@tiptap/extension-code-block-lowlight`, `lowlight` and `@tiptap/markdown` (official) for the editor. Markdown is the storage format; the editor loads and serializes markdown. If `@tiptap/markdown` is not installable, use `tiptap-markdown` and record the choice here. Installed: `@tiptap/markdown` (official, 3.x). `@tiptap/core` and `@tiptap/pm` are listed as direct dependencies because they are peer dependencies of `@tiptap/react` and the extensions. Mentions serialize as `@[Name](user:<id>)` through a `Mention.extend` with its own markdown tokenizer (`components/shared/markdown-mention.ts`).
- `react-markdown` + `remark-gfm` for read-only rendering where an editor is too heavy (notification snippets, previews)
- `emoji-picker-react` for reactions and icons

## Theme

`src/styles/tokens.css` keeps Unbind's variable names (`--background`, `--foreground`, `--card`, `--primary`, `--muted-foreground`, `--border`, `--input`, `--ring`, `--destructive`, `--warning`, `--success`, `--process`, `--change`, chart colors, shadows, radius) so every ported component works unchanged. Palette: neutral with a blue accent. Light: background `hsl(220 20% 98%)`, foreground `hsl(220 25% 8%)`, card `hsl(0 0% 100%)`, border `hsl(220 14% 90%)`, primary `hsl(235 60% 55%)`, primary-foreground white, muted-foreground `hsl(220 10% 42%)`. Dark: background `hsl(222 16% 9%)`, foreground `hsl(220 15% 93%)`, card `hsl(222 14% 12%)`, border `hsl(222 12% 20%)`, primary `hsl(235 70% 68%)`. Priority colors: urgent `--destructive`, high `--warning`, medium `--process`, low `--muted-foreground`. State type colors come from the state record.

## Routes (TanStack file routes, `src/routes`)

```
/sign-in                                 email + password; link to forgot password
/forgot-password, /reset-password        password reset flow
/welcome                                 first user + first workspace (only while no user exists)
/invite/$token                           accept invitation
/oauth/consent                           connected app consent (port from Unbind)
/                                        redirect to last workspace or first
/$workspace/_workspace                   app shell: sidebar + outlet (workspace = slug)
  /inbox                                 notifications, split view with the selected issue
  /my-issues                             tabs: assigned, created, subscribed; list
  /issues                                all issues in the workspace (list/board)
  /issue/$identifier                     issue detail page (full page); also opens as a side peek from lists
  /projects                              projects list/board/timeline-lite (list and board only)
  /project/$slug                         project overview (description, updates, milestones) and issues tab: /project/$slug/issues
  /initiatives, /initiative/$slug
  /views, /view/$view_id
  /documents, /document/$slug
  /team/$key                             redirect to /team/$key/issues
  /team/$key/issues                      active issues (list/board), filters, display options, grouped by state
  /team/$key/backlog
  /team/$key/triage                      when triage enabled
  /team/$key/cycles, /team/$key/cycle/$number
  /team/$key/projects
  /team/$key/labels
  /settings                              workspace settings (admin): General, Members, Teams, Labels, Project statuses, Templates, Webhooks, Import, Danger zone
  /settings/team/$key                    team settings: General, Workflow, Cycles, Estimates, Triage, Members, Labels, Templates, Auto close & archive, Danger zone
  /settings/account                      Profile, Password, Notifications, API keys, Connected apps
  /settings/system                       instance (system admin): General, Email, Users
```

Everything under `$workspace` is guarded by `whoami`; an unknown slug redirects to the first workspace; a user with no workspace sees `/create-workspace`.

## Shell

Left sidebar (width 240px desktop, drawer below `md`): workspace switcher (name, icon, dropdown: other workspaces, settings, create workspace, sign out), search button (opens command panel, `cmd+k`), "New issue" button (`c`), then nav: Inbox (unread badge), My issues; section "Workspace": Projects, Views, Initiatives, Documents, All issues; section "Favorites"; section "Your teams": each team expands to Issues, Backlog, Triage (if enabled), Cycles (if enabled), Projects; footer: settings gear, help. Top of the content area: a `PageHeader` with breadcrumb and page actions.

Command panel (`cmdk`): search issues, projects, documents, teams; actions: new issue, go to inbox, go to my issues, go to settings, switch workspace, toggle theme.

## Layout primitives (`src/components/layout`)

Use these and nothing else for page structure:

- `PageContainer` the scrolling content column, `px-6 py-5` on desktop, `px-4 py-4` below `sm`; `max-w-none` for lists, `max-w-3xl` for documents and settings via `width="narrow"`.
- `PageHeader` title row: `h1` `text-xl font-semibold leading-tight`, optional icon, breadcrumb (`Breadcrumb` component), right slot for actions, below it an optional `PageToolbar`.
- `PageToolbar` the filter/display row: height 40px, `gap-2`, left slot filters, right slot display options and view switch.
- `SettingsPage` (title, description, sections) and `SettingsSection` (title, description, children, optional footer with save button): port from Unbind. Sections are `gap-6` apart; fields inside a section `gap-4`.
- `EmptyState` icon, title, description, optional action button. Used for every empty list.
- `ErrorCard` message plus retry. Used for every failed query.
- `ListSkeleton` rows of placeholders, used for every pending list.
- `SplitView` for inbox and list+peek: resizable, right pane 480px default, hidden below `lg`.

## Spacing and type scale (enforced)

Only these utilities for the listed purposes. Do not use arbitrary values (`p-[13px]`) anywhere.

| Purpose | Utility |
| --- | --- |
| Gap between icon and label | `gap-1.5` |
| Gap between inline controls in a row | `gap-2` |
| Gap between list rows | none; rows are `h-9` (dense) with `border-b` |
| Gap between cards in a grid | `gap-3` |
| Gap between form fields | `gap-4` |
| Gap between page sections | `gap-6` |
| Card padding | `p-4` |
| Dialog padding | from `DialogContent` only |
| Page title | `text-xl font-semibold leading-tight` via `PageHeader` |
| Section title | `text-base font-semibold` via `SettingsSection` or `SectionTitle` |
| Body | default `text-sm` |
| Secondary text | `text-sm text-muted-foreground` |
| Captions, metadata, identifiers | `text-xs text-muted-foreground` |
| Monospace identifiers | `font-mono text-xs` |
| Radius | `rounded-md` for controls and rows, `rounded-lg` for cards and dialogs, `rounded-full` for avatars and pills |

Buttons: `Button` from `components/ui/button` only, sizes `sm` (28px, used in toolbars, rows, dialogs' footers), `default` (36px, page actions and forms), `icon` (32px). Never style a `<button>` by hand. Primary action per dialog is `variant="default"`, secondary `variant="outline"`, destructive `variant="destructive"`. Link-like actions use `variant="ghost"`.

Inputs: `Input`, `Textarea`, `Select`/`Combobox`, `Checkbox`, `Switch` from `components/ui` only. Form fields render through `useAppForm` field components (`AppField.TextField`, `.TextareaField`, `.SelectField`, `.SwitchField`) with label, description and error in one shape.

Avatars: `Avatar` component, sizes `xs` 16, `sm` 20, `md` 24, `lg` 32, `xl` 64; initials fallback with a color derived from the user id.

Icons: lucide only, `size-4` default in text and rows, `size-3.5` in `text-xs` contexts, `size-5` in page headers. Priority, state, estimate and label icons are shared components: `PriorityIcon`, `StateIcon` (circle variants per state type: dashed for backlog, empty for unstarted, partial for started, check for completed, x for canceled, triangle for triage), `LabelChip`, `EstimateBadge`, `ProjectIcon`, `CycleIcon`, `TeamIcon`.

Dates: `date-fns`; relative ("3h ago") in rows and activity, absolute (`MMM d, yyyy`) in detail properties, with a tooltip holding the full timestamp. One helper `formatRelative` and one `formatDate` in `lib/helpers/dates.ts`.

Toasts: `toast` from `components/ui/toast` for every mutation failure, and for successes that happen off screen (archived, moved, copied link). No `alert()`, no inline success text.

## Issue list and board

`IssueList`: virtualized rows (`@tanstack/react-virtual`), each row: priority icon, identifier (`font-mono text-xs text-muted-foreground w-16`), state icon, title (truncate), label chips, project chip, due date, assignee avatar, right aligned. Rows group by state by default with a group header (state icon, name, count, "+" to create in that state). Clicking opens the side peek on `lg+`, navigates on smaller. `cmd/ctrl+click` opens the full page. Selection with click + shift/cmd and `x`; selected rows show a bulk action bar at the bottom (`BulkActionBar`): state, priority, assignee, labels, project, cycle, archive, trash.

`IssueBoard`: columns per state (or per group), cards: identifier, title, labels, assignee, priority. Drag between columns updates state; drag within a column updates `sort_order` (midpoint between neighbors, re-spaced by the API when gaps get small).

`IssueFilters` (toolbar): "Filter" button opens a menu of fields (status, assignee, creator, priority, labels, project, cycle, due date, subscriber); each applied filter is a chip with an inline operator and value picker; filters serialize to search params with the same names as `list-issues`. `DisplayOptions` popover: layout (list/board), grouping, ordering, show sub-issues, show completed (`all|past week|none`), visible properties. Views save the current filters and display as a `View`.

`IssueProperties` (right column in detail, same component in the side peek and the new issue dialog): State, Priority, Assignee, Labels, Project, Milestone, Cycle, Estimate, Due date, Parent, Subscribers. Each property is a `PropertyRow` with a `PropertyPicker` (`Popover` + `Command` list, keyboard searchable). The same pickers back the keyboard shortcuts: `s` state, `p` priority, `a` assignee, `l` labels, `shift+p` project, `shift+c` cycle, `e` estimate, `d` due date.

`NewIssueDialog`: title input (autofocus), description editor, properties row (state, priority, assignee, labels, project, cycle, estimate, due date, parent) as compact pickers, team picker on the left of the title, "Create more" toggle, submit `cmd+enter`. Opens with `c` from anywhere in the workspace and from every "+" button; defaults from the current context (team, state, project, cycle, parent, filters).

`IssueDetail`: header (breadcrumb team > identifier, actions: copy link, copy branch name, subscribe, archive, trash, open in full page), title (inline editable), description editor (autosave with 800ms debounce, saved indicator), sub-issues list with "Add sub-issue", attachments, activity feed (history entries and comments interleaved by time; comments are threaded, each with reactions, edit, delete, resolve), comment composer at the bottom (`cmd+enter`), properties column on the right.

Issue slice notes (decisions the paragraphs above do not cover):

- Rows carry a selection checkbox in a left gutter, visible on hover and while a selection exists; shift+click on a row extends the selection from the last toggled row, `x` toggles the hovered row, `Escape` clears.
- The side peek is driven by a `peek=<identifier>` search param through `IssuePeekHost` (`components/issues/issue-peek.tsx`); lists rendered outside a host navigate to the full page instead. Embedders wrap their page in `IssuePeekHost` to get the peek.
- Filter chips show `is` / `is any of` without a picker for list fields, because `list-issues` has no negation; the due date chip has an operator picker (`before`, `after`).
- `show_completed: past_week` is applied client side on the fetched pages; `none` narrows `state_types` in the request.
- Team issues page fixes `state_types` to unstarted, started, completed and canceled; backlog to backlog; triage to triage.
- The detail title is an inline editable `Textarea` with the page title style (`text-xl font-semibold leading-tight`), the one place that style is used outside `PageHeader`, because the title must be editable in place.
- In the detail, the sub-issues and attachments sections show only their header and add actions when empty, not an `EmptyState`, which would dominate the page.
- `PopoverContent` is layered `z-[1000]` like `Select` and `DropdownMenu`, so pickers open above dialogs.

Projects: `ProjectList` (grouped by status; columns: name, status, lead, members, target date, progress bar, priority), `ProjectBoard`, `ProjectDetail` with Overview (summary, description editor, properties: status, lead, members, teams, dates, priority, labels, links; milestones with progress; updates timeline with health badges and composer) and Issues tabs (reuses `IssueList` with `project_ids` fixed).

Projects slice notes (decisions the paragraph above does not cover):

- `ProjectList` and `ProjectBoard` (`components/projects`) take `{workspaceId, filters, emptyState?}` with `filters` named like `list-projects`; the projects page, the team projects page and the initiative page all render them. The group header and the board column copy the issue list's header and column styles.
- Project, initiative and document pages share one detail layout: a breadcrumb row with icon actions (favorite, copy link, archive, trash, delete), the `LifecycleBanner` from `components/documents/lifecycle-banner.tsx`, an `IconPicker` next to the inline editable name (`InlineText` from `components/projects/inline-text.tsx`, the same `Textarea` style as the issue title), the description through `AutosaveMarkdown` (`components/projects/autosave-markdown.tsx`, which skips a save when the markdown did not change), and for projects and initiatives the properties column at `lg:w-72`. The project's Overview and Issues tabs sit under the breadcrumb row; the Issues tab is `IssuesView` with `project_ids` fixed, so its new issue button defaults the project.
- Project and initiative status updates share `StatusUpdateComposer` and `StatusUpdateCard` (`components/projects`), with `HealthBadge`; project updates add the issue `ReactionBar`.
- Project dates show at their resolution (`Oct 16, 2026`, `Oct 2026`, `Q4 2026`, `H2 2026`, `2026`); picking a coarser resolution moves the date to the last day of its period.
- Project statuses use `ProjectStatusIcon`, which maps backlog, planned, started, paused, completed and canceled onto the `StateIcon` shapes.

Cycles: `CycleList` (active, upcoming, past), `CycleDetail` (header with dates, progress bar and scope numbers, burn-up bars from `scope_history` drawn as plain SVG since no chart library is listed, issue list filtered by cycle).

Favorites: `FavoriteButton({kind, entityId})` in page headers; the sidebar Favorites section is `components/favorites/sidebar-favorites.tsx`, drag to reorder. Saved views: `useSaveViewDialog({filters, display, teamId?})` from `components/views/save-view-dialog.tsx` returns `{openSaveView, saveViewDialog}`; a view stores `show_completed` as a bool, so `past_week` is saved as shown.

Inbox: `NotificationList` (tabs: Inbox, Read, Archived, Snoozed; rows: actor avatar, title, issue identifier, time; unread dot) with the selected notification's issue in the `SplitView` right pane. Keyboard: `j/k` move, `e` archive, `u` unread, `h` snooze.

Settings pages all use `SettingsPage` + `SettingsSection`. Members page: table rows (avatar, name, email, role select, teams, remove); invite dialog (email, role, teams); pending invitations list with resend/revoke/copy link. Labels page: groups collapsible, drag to reorder, color picker (12 preset swatches + hex input). Workflow page: states by type with drag to reorder, add state, delete state with "move issues to" picker. Import page: upload JSON, options (dry run, restore archived, restore trashed, create missing users), run, show report with counts, warnings and user mapping; a second "Run for real" button after a dry run.

## Foundation notes

Decisions made while building the foundation that the sections above do not cover:

- `Button` sizes: `sm` is `h-7`, `default` is `h-9`, `icon` is `size-8`. Unbind's `lg` size is removed so it cannot be used.
- Extra layout primitives: `List`, `ListRow`, `ListRowText` (bordered list of `h-9` rows), `ListStates` (pending, error and empty handling for one list in one component), `SettingsLayout` (settings nav plus a narrow `PageContainer`), `AuthShell` (centered public pages) and `NotBuiltYet` (the placeholder empty state).
- Extra `components/ui`: `file-trigger` (the only hidden file input), `suggestion-menu` (floating list for editor mentions) and `resizable` (written in house; there is no resizable primitive in base-ui and no library is listed for it).
- Settings routes: the workspace tabs live under a pathless `settings/_admin` layout so `settings/team/$key`, `settings/account` and `settings/system` keep their own layouts. The settings nav is vertical from `lg` and a horizontal scroll row below. Placeholder settings tabs render `SettingsPage` + `NotBuiltYet`, because `SettingsLayout` already provides the `PageContainer`.
- `/$workspace` redirects to `/$workspace/my-issues`.
- Mutation failures toast from one place: the `QueryClient`'s `MutationCache.onError` reads `meta.errorTitle` from each mutation hook. `meta.silent` opts out.
- Forms submit through `submitHandler(form)` from `use-app-form`, which swallows the rethrown submit error because the toast already reported it.
- Types in query modules are derived from the generated client's method signatures (`TResponseData`, `TListItem`, `TInput` in `lib/server/types.ts`), not from schema names, so the Go struct names do not leak into the web code.
- `scripts/generate-sdk`: a path served by more than one method (for example `GET` and `PUT /users/notification-settings`) generates one child per method, `client.users.notificationSettings.get()` and `.put()`. Multipart request bodies (`create-upload`) take a `FormData`.
- Command panel search results come from `useCommandPanelSearch(query)` in `components/command-panel/use-command-panel-search.ts`, which returns no groups until search is wired.
- `New issue` (`c` and the sidebar button) opens `NewIssueDialog` from `components/issues`, a placeholder driven by `useNewIssueStore` with context defaults.
- Dark theme `--primary-foreground` is the dark background color, not white, because white on `hsl(235 70% 68%)` has low contrast.
- `scripts/check-ui-consistency.mjs` (part of `npm run lint`) fails on arbitrary Tailwind values (`p-[13px]`, `w-(--x)`) outside `components/ui` and `components/layout`, raw `<button>`/`<input>` outside `components/ui`, `text-2xl` and larger outside `components/layout`, and `alert(`.

## Query modules (`src/lib/queries/*`)

One file per API group mirroring Unbind's pattern: `queryKey<Group>` object, `*Query` functions returning `queryOptions`, plain async mutation helpers, and `use<Group>Mutation` hooks that invalidate the right keys. Lists invalidate on every mutation of their entity; issue mutations also invalidate `notifications.counts`. Default `staleTime` 5s; list pages `refetchInterval` 15s when the tab is visible.

## Accessibility

Every icon-only button has `aria-label`. Pickers are keyboard navigable (base-ui). Focus rings are the `focus-visible` styles from the ported components. Color is never the only signal: state and priority have icons.
