import type { TFieldChange } from "@/lib/queries/issues";

const PRIORITY_NAMES: Record<string, string> = {
  "0": "No priority",
  "1": "Urgent",
  "2": "High",
  "3": "Medium",
  "4": "Low",
};

const FIELD_NAMES: Record<string, string> = {
  project: "project",
  milestone: "milestone",
  cycle: "cycle",
  parent: "parent issue",
  due_date: "due date",
  estimate: "estimate",
};

function valueOf(label: string | undefined, raw: unknown, field: string) {
  const value = label || (raw === undefined || raw === null ? "" : String(raw));
  if (field === "priority") return PRIORITY_NAMES[value] ?? value;
  return value;
}

export function describeChange(change: TFieldChange): string {
  const from = valueOf(change.from_label, change.from, change.field);
  const to = valueOf(change.to_label, change.to, change.field);
  switch (change.field) {
    case "created":
      return "created the issue";
    case "imported":
      return "imported the issue";
    case "title":
      return `changed the title to "${to}"`;
    case "description":
      return "updated the description";
    case "state":
      return from ? `changed status from ${from} to ${to}` : `set status to ${to}`;
    case "priority":
      return from ? `changed priority from ${from} to ${to}` : `set priority to ${to}`;
    case "assignee":
      if (!to) return from ? `unassigned ${from}` : "removed the assignee";
      return `assigned the issue to ${to}`;
    case "labels_added":
      return `added label ${to}`;
    case "labels_removed":
      return `removed label ${from || to}`;
    case "relation_added":
      return `added relation ${to}`;
    case "relation_removed":
      return `removed relation ${from || to}`;
    case "attachment_added":
      return `added attachment ${to}`;
    case "attachment_removed":
      return `removed attachment ${from || to}`;
    case "github":
      return change.to_label ? `on GitHub: ${change.to_label}` : "updated from GitHub";
    case "archived":
      return "archived the issue";
    case "unarchived":
      return "unarchived the issue";
    case "trashed":
      return "moved the issue to trash";
    case "restored":
      return "restored the issue";
    case "team":
      return `moved the issue from ${from} to ${to}`;
    default: {
      const name = FIELD_NAMES[change.field] ?? change.field.replace(/_/g, " ");
      if (!to) return `removed the ${name}`;
      return `set ${name} to ${to}`;
    }
  }
}
