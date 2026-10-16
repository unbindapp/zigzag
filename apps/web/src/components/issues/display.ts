import type { TGroupBy, TOrderBy } from "@/lib/queries/issues";

export type TIssueLayout = "list" | "board";
export type TShowCompleted = "all" | "past_week" | "none";
export type TIssueProperty =
  | "priority"
  | "identifier"
  | "state"
  | "labels"
  | "project"
  | "cycle"
  | "estimate"
  | "due_date"
  | "assignee"
  | "updated_at";

export type TDisplaySettings = {
  layout: TIssueLayout;
  group_by: TGroupBy;
  order_by: TOrderBy;
  order: "asc" | "desc";
  nest_sub_issues: boolean;
  show_completed: TShowCompleted;
  properties: TIssueProperty[];
};

export type DisplaySettings = TDisplaySettings;

export const ISSUE_PROPERTIES: { value: TIssueProperty; label: string }[] = [
  { value: "priority", label: "Priority" },
  { value: "identifier", label: "ID" },
  { value: "state", label: "Status" },
  { value: "labels", label: "Labels" },
  { value: "project", label: "Project" },
  { value: "cycle", label: "Cycle" },
  { value: "estimate", label: "Estimate" },
  { value: "due_date", label: "Due date" },
  { value: "assignee", label: "Assignee" },
  { value: "updated_at", label: "Updated" },
];

export const GROUP_BY_OPTIONS: { value: TGroupBy; label: string }[] = [
  { value: "state", label: "Status" },
  { value: "assignee", label: "Assignee" },
  { value: "priority", label: "Priority" },
  { value: "project", label: "Project" },
  { value: "cycle", label: "Cycle" },
  { value: "label", label: "Label" },
  { value: "parent", label: "Parent issue" },
  { value: "none", label: "No grouping" },
];

export const ORDER_BY_OPTIONS: { value: TOrderBy; label: string }[] = [
  { value: "manual", label: "Manual" },
  { value: "priority", label: "Priority" },
  { value: "updated_at", label: "Last updated" },
  { value: "created_at", label: "Last created" },
  { value: "due_date", label: "Due date" },
  { value: "title", label: "Title" },
];

export const SHOW_COMPLETED_OPTIONS: { value: TShowCompleted; label: string }[] = [
  { value: "all", label: "All" },
  { value: "past_week", label: "Past week" },
  { value: "none", label: "None" },
];

export const DEFAULT_DISPLAY: TDisplaySettings = {
  layout: "list",
  group_by: "state",
  order_by: "manual",
  order: "asc",
  nest_sub_issues: true,
  show_completed: "past_week",
  properties: ["priority", "identifier", "state", "labels", "project", "due_date", "assignee"],
};

export function defaultOrderFor(orderBy: TOrderBy): "asc" | "desc" {
  switch (orderBy) {
    case "created_at":
    case "updated_at":
      return "desc";
    default:
      return "asc";
  }
}

export function resolveDisplay(
  ...layers: (Partial<TDisplaySettings> | undefined)[]
): TDisplaySettings {
  return layers.reduce<TDisplaySettings>((acc, layer) => ({ ...acc, ...layer }), DEFAULT_DISPLAY);
}
