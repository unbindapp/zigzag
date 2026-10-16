import type {
  TGroupBy,
  TIssue,
  TLabelSummary,
  TOrderBy,
  TUserSummary,
  TWorkflowStateSummary,
} from "@/lib/queries/issues";
import type { TStateType } from "@/components/shared/state-icon";

export const STATE_TYPE_ORDER: TStateType[] = [
  "triage",
  "backlog",
  "unstarted",
  "started",
  "completed",
  "canceled",
];

export const PRIORITY_ORDER = [1, 2, 3, 4, 0];

const PRIORITY_NAMES: Record<number, string> = {
  0: "No priority",
  1: "Urgent",
  2: "High",
  3: "Medium",
  4: "Low",
};

export type TGroupMeta =
  | { kind: "none" }
  | { kind: "state"; state: TWorkflowStateSummary }
  | { kind: "priority"; priority: number }
  | { kind: "assignee"; user: TUserSummary | null }
  | { kind: "project"; project: NonNullable<TIssue["project"]> | null }
  | { kind: "cycle"; cycle: NonNullable<TIssue["cycle"]> | null }
  | { kind: "label"; label: TLabelSummary | null }
  | { kind: "parent"; parent: NonNullable<TIssue["parent"]> | null };

export type TIssueGroup = {
  key: string;
  label: string;
  meta: TGroupMeta;
  issues: TIssue[];
};

export const NONE_KEY = "none";

export function compareStates(a: TWorkflowStateSummary, b: TWorkflowStateSummary) {
  const byType = STATE_TYPE_ORDER.indexOf(a.type) - STATE_TYPE_ORDER.indexOf(b.type);
  if (byType !== 0) return byType;
  return a.position - b.position;
}

export function sortStates<T extends TWorkflowStateSummary>(states: T[]) {
  return [...states].sort(compareStates);
}

function compareNullableDates(a: string | null, b: string | null) {
  if (a === b) return 0;
  if (!a) return 1;
  if (!b) return -1;
  return a.localeCompare(b);
}

export function compareIssues(orderBy: TOrderBy, order: "asc" | "desc") {
  const direction = order === "asc" ? 1 : -1;
  return (a: TIssue, b: TIssue) => {
    switch (orderBy) {
      case "manual":
        return direction * (a.sort_order - b.sort_order);
      case "priority":
        return (
          direction * (PRIORITY_ORDER.indexOf(a.priority) - PRIORITY_ORDER.indexOf(b.priority))
        );
      case "created_at":
        return direction * a.created_at.localeCompare(b.created_at);
      case "updated_at":
        return direction * a.updated_at.localeCompare(b.updated_at);
      case "due_date":
        return direction * compareNullableDates(a.due_date, b.due_date);
      case "title":
        return direction * a.title.localeCompare(b.title);
    }
  };
}

function stateGroupKey(state: TWorkflowStateSummary, byId: boolean) {
  return byId ? state.id : `${state.type}:${state.name.toLowerCase()}`;
}

type TGroupSeed = { key: string; label: string; meta: TGroupMeta; rank: number | string };

function seedsFor(issue: TIssue, groupBy: TGroupBy, statesById: boolean): TGroupSeed[] {
  switch (groupBy) {
    case "none":
      return [{ key: "all", label: "All issues", meta: { kind: "none" }, rank: 0 }];
    case "state":
      return [
        {
          key: stateGroupKey(issue.state, statesById),
          label: issue.state.name,
          meta: { kind: "state", state: issue.state },
          rank: STATE_TYPE_ORDER.indexOf(issue.state.type) * 10000 + issue.state.position,
        },
      ];
    case "priority":
      return [
        {
          key: String(issue.priority),
          label: PRIORITY_NAMES[issue.priority] ?? PRIORITY_NAMES[0],
          meta: { kind: "priority", priority: issue.priority },
          rank: PRIORITY_ORDER.indexOf(issue.priority),
        },
      ];
    case "assignee": {
      const user = issue.assignee ?? null;
      return [
        {
          key: user?.id ?? NONE_KEY,
          label: user ? user.name || user.display_name : "No assignee",
          meta: { kind: "assignee", user },
          rank: user ? `0${(user.name || user.display_name).toLowerCase()}` : "1",
        },
      ];
    }
    case "project": {
      const project = issue.project ?? null;
      return [
        {
          key: project?.id ?? NONE_KEY,
          label: project?.name ?? "No project",
          meta: { kind: "project", project },
          rank: project ? `0${project.name.toLowerCase()}` : "1",
        },
      ];
    }
    case "cycle": {
      const cycle = issue.cycle ?? null;
      return [
        {
          key: cycle?.id ?? NONE_KEY,
          label: cycle ? cycle.name || `Cycle ${cycle.number}` : "No cycle",
          meta: { kind: "cycle", cycle },
          rank: cycle ? `0${String(cycle.number).padStart(6, "0")}` : "1",
        },
      ];
    }
    case "label": {
      if (issue.labels.length === 0) {
        return [
          { key: NONE_KEY, label: "No label", meta: { kind: "label", label: null }, rank: "1" },
        ];
      }
      return issue.labels.map((label) => ({
        key: label.id,
        label: label.name,
        meta: { kind: "label", label },
        rank: `0${label.name.toLowerCase()}`,
      }));
    }
    case "parent": {
      const parent = issue.parent ?? null;
      return [
        {
          key: parent?.id ?? NONE_KEY,
          label: parent ? `${parent.identifier} ${parent.title}` : "No parent issue",
          meta: { kind: "parent", parent },
          rank: parent ? `0${parent.identifier}` : "1",
        },
      ];
    }
  }
}

function compareRanks(a: number | string, b: number | string) {
  if (typeof a === "number" && typeof b === "number") return a - b;
  return String(a).localeCompare(String(b));
}

export function groupIssues({
  issues,
  groupBy,
  orderBy,
  order,
  states,
}: {
  issues: TIssue[];
  groupBy: TGroupBy;
  orderBy: TOrderBy;
  order: "asc" | "desc";
  states?: TWorkflowStateSummary[];
}): TIssueGroup[] {
  const statesById = Boolean(states);
  const groups = new Map<string, { group: TIssueGroup; rank: number | string }>();
  const ensure = (seed: TGroupSeed) => {
    const existing = groups.get(seed.key);
    if (existing) return existing.group;
    const group: TIssueGroup = { key: seed.key, label: seed.label, meta: seed.meta, issues: [] };
    groups.set(seed.key, { group, rank: seed.rank });
    return group;
  };
  if (groupBy === "state" && states) {
    for (const state of sortStates(states)) {
      ensure({
        key: state.id,
        label: state.name,
        meta: { kind: "state", state },
        rank: STATE_TYPE_ORDER.indexOf(state.type) * 10000 + state.position,
      });
    }
  }
  if (groupBy === "priority") {
    for (const priority of PRIORITY_ORDER) {
      ensure({
        key: String(priority),
        label: PRIORITY_NAMES[priority],
        meta: { kind: "priority", priority },
        rank: PRIORITY_ORDER.indexOf(priority),
      });
    }
  }
  for (const issue of issues) {
    for (const seed of seedsFor(issue, groupBy, statesById)) {
      ensure(seed).issues.push(issue);
    }
  }
  const compare = compareIssues(orderBy, order);
  return [...groups.values()]
    .sort((a, b) => compareRanks(a.rank, b.rank))
    .map(({ group }) => ({ ...group, issues: group.issues.sort(compare) }));
}
