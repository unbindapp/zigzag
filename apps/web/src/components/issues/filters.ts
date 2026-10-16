import { z } from "zod";

import type { TStateType } from "@/components/shared/state-icon";
import type { TListIssuesInput } from "@/lib/queries/issues";
import type { TDisplaySettings } from "@/components/issues/display";

export const UNASSIGNED = "null";

export type ListIssuesFilters = {
  team_ids?: string[];
  state_ids?: string[];
  state_types?: TStateType[];
  assignee_ids?: string[];
  creator_ids?: string[];
  label_ids?: string[];
  project_ids?: string[];
  milestone_ids?: string[];
  cycle_ids?: string[];
  parent_id?: string;
  priorities?: number[];
  due_before?: string;
  due_after?: string;
  subscriber_id?: string;
  query?: string;
};

export type TFilterField =
  | "status"
  | "assignee"
  | "creator"
  | "priority"
  | "labels"
  | "project"
  | "cycle"
  | "due_date"
  | "subscriber";

export const FILTER_FIELDS: { field: TFilterField; label: string }[] = [
  { field: "status", label: "Status" },
  { field: "assignee", label: "Assignee" },
  { field: "creator", label: "Creator" },
  { field: "priority", label: "Priority" },
  { field: "labels", label: "Labels" },
  { field: "project", label: "Project" },
  { field: "cycle", label: "Cycle" },
  { field: "due_date", label: "Due date" },
  { field: "subscriber", label: "Subscriber" },
];

type TListParam =
  "state_ids" | "assignee_ids" | "creator_ids" | "label_ids" | "project_ids" | "cycle_ids";

export const listParamOfField: Partial<Record<TFilterField, TListParam>> = {
  status: "state_ids",
  assignee: "assignee_ids",
  creator: "creator_ids",
  labels: "label_ids",
  project: "project_ids",
  cycle: "cycle_ids",
};

const STATE_TYPE_VALUES = [
  "triage",
  "backlog",
  "unstarted",
  "started",
  "completed",
  "canceled",
] as const satisfies readonly TStateType[];

const COMPLETED_TYPES: TStateType[] = ["completed", "canceled"];

const stringList = z.array(z.string()).optional().catch(undefined);
const optionalString = z.string().optional().catch(undefined);
const datePattern = /^\d{4}-\d{2}-\d{2}$/;
const optionalDate = z.string().regex(datePattern).optional().catch(undefined);

export const issueSearchSchema = z.object({
  state_ids: stringList,
  state_types: z.array(z.enum(STATE_TYPE_VALUES)).optional().catch(undefined),
  assignee_ids: stringList,
  creator_ids: stringList,
  label_ids: stringList,
  project_ids: stringList,
  cycle_ids: stringList,
  priorities: z.array(z.number().int().min(0).max(4)).optional().catch(undefined),
  due_before: optionalDate,
  due_after: optionalDate,
  subscriber_id: optionalString,
  peek: optionalString,
});

export type TIssueSearch = z.infer<typeof issueSearchSchema>;

export function filtersFromSearch(search: TIssueSearch): ListIssuesFilters {
  const filters: ListIssuesFilters = {
    state_ids: search.state_ids,
    state_types: search.state_types,
    assignee_ids: search.assignee_ids,
    creator_ids: search.creator_ids,
    label_ids: search.label_ids,
    project_ids: search.project_ids,
    cycle_ids: search.cycle_ids,
    priorities: search.priorities,
    due_before: search.due_before,
    due_after: search.due_after,
    subscriber_id: search.subscriber_id,
  };
  return compactFilters(filters);
}

export function compactFilters<T extends object>(filters: T): T {
  const entries = Object.entries(filters).filter(([, value]) => {
    if (value === undefined || value === null || value === "") return false;
    if (Array.isArray(value)) return value.length > 0;
    return true;
  });
  return Object.fromEntries(entries) as T;
}

export function filtersToSearch(filters: ListIssuesFilters): Partial<TIssueSearch> {
  const clean = compactFilters(filters);
  return {
    state_ids: clean.state_ids,
    state_types: clean.state_types,
    assignee_ids: clean.assignee_ids,
    creator_ids: clean.creator_ids,
    label_ids: clean.label_ids,
    project_ids: clean.project_ids,
    cycle_ids: clean.cycle_ids,
    priorities: clean.priorities,
    due_before: clean.due_before,
    due_after: clean.due_after,
    subscriber_id: clean.subscriber_id,
  };
}

export function activeFilterFields(filters: ListIssuesFilters): TFilterField[] {
  const clean = compactFilters(filters);
  return FILTER_FIELDS.map((item) => item.field).filter((field) => {
    switch (field) {
      case "status":
        return Boolean(clean.state_ids || clean.state_types);
      case "priority":
        return Boolean(clean.priorities);
      case "due_date":
        return Boolean(clean.due_before || clean.due_after);
      case "subscriber":
        return Boolean(clean.subscriber_id);
      default: {
        const param = listParamOfField[field];
        return Boolean(param && clean[param]);
      }
    }
  });
}

export function clearFilterField(filters: ListIssuesFilters, field: TFilterField) {
  const next = { ...filters };
  switch (field) {
    case "status":
      delete next.state_ids;
      delete next.state_types;
      break;
    case "priority":
      delete next.priorities;
      break;
    case "due_date":
      delete next.due_before;
      delete next.due_after;
      break;
    case "subscriber":
      delete next.subscriber_id;
      break;
    default: {
      const param = listParamOfField[field];
      if (param) delete next[param];
    }
  }
  return compactFilters(next);
}

function intersect<T>(a: T[] | undefined, b: T[] | undefined): T[] | undefined {
  if (!a) return b;
  if (!b) return a;
  return a.filter((value) => b.includes(value));
}

export function mergeFilters(
  filters: ListIssuesFilters,
  fixed: ListIssuesFilters | undefined,
): ListIssuesFilters {
  if (!fixed) return compactFilters(filters);
  const merged: ListIssuesFilters = { ...filters, ...fixed };
  const listKeys = [
    "team_ids",
    "state_ids",
    "state_types",
    "assignee_ids",
    "creator_ids",
    "label_ids",
    "project_ids",
    "milestone_ids",
    "cycle_ids",
  ] as const;
  for (const key of listKeys) {
    const value = intersect<string>(filters[key], fixed[key]);
    (merged as Record<string, unknown>)[key] = value;
  }
  merged.priorities = intersect(filters.priorities, fixed.priorities);
  return merged;
}

function stateTypesForDisplay(
  current: TStateType[] | undefined,
  showCompleted: TDisplaySettings["show_completed"],
): TStateType[] | undefined {
  if (showCompleted !== "none") return current;
  const base = current ?? [...STATE_TYPE_VALUES];
  return base.filter((type) => !COMPLETED_TYPES.includes(type));
}

export function buildListIssuesInput({
  workspaceId,
  filters,
  fixedFilters,
  display,
}: {
  workspaceId: string;
  filters: ListIssuesFilters;
  fixedFilters?: ListIssuesFilters;
  display: TDisplaySettings;
}): TListIssuesInput {
  const merged = mergeFilters(filters, fixedFilters);
  const input: TListIssuesInput = {
    ...merged,
    workspace_id: workspaceId,
    state_types: stateTypesForDisplay(merged.state_types, display.show_completed),
    order_by: display.order_by,
    order: display.order,
    group_by: display.group_by,
  };
  return compactFilters(input);
}

const WEEK_MS = 7 * 24 * 60 * 60 * 1000;

export function isHiddenCompleted(
  issue: { state: { type: TStateType }; completed_at: string | null; canceled_at: string | null },
  showCompleted: TDisplaySettings["show_completed"],
  now: number,
) {
  if (!COMPLETED_TYPES.includes(issue.state.type)) return false;
  switch (showCompleted) {
    case "all":
      return false;
    case "none":
      return true;
    case "past_week": {
      const closedAt = issue.completed_at ?? issue.canceled_at;
      if (!closedAt) return false;
      return now - new Date(closedAt).getTime() > WEEK_MS;
    }
  }
}
