import { parseISO } from "date-fns";
import { CalendarIcon } from "lucide-react";

import type { TIssue } from "@/lib/queries/issues";
import { ProjectIcon } from "@/components/shared/project-icon";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { formatDate } from "@/lib/helpers/dates";

export function ProjectChip({ project }: { project: NonNullable<TIssue["project"]> }) {
  return (
    <span className="bg-background text-foreground inline-flex h-5 max-w-40 shrink-0 items-center gap-1.5 rounded-full border px-2 text-xs leading-none">
      <ProjectIcon icon={project.icon} color={project.color} size="sm" />
      <span className="truncate">{project.name}</span>
    </span>
  );
}

function isOverdue(dueDate: string, isClosed: boolean) {
  if (isClosed) return false;
  const due = parseISO(dueDate);
  due.setHours(23, 59, 59, 999);
  return due.getTime() < Date.now();
}

export function DueDateText({ issue }: { issue: Pick<TIssue, "due_date" | "state"> }) {
  if (!issue.due_date) return null;
  const isClosed = issue.state.type === "completed" || issue.state.type === "canceled";
  const overdue = isOverdue(issue.due_date, isClosed);
  const text = formatDate(parseISO(issue.due_date));
  return (
    <SimpleTooltip content={overdue ? `Overdue, due ${text}` : `Due ${text}`}>
      <span
        data-overdue={overdue || undefined}
        className="text-muted-foreground data-overdue:text-destructive inline-flex shrink-0 items-center gap-1.5 text-xs whitespace-nowrap"
      >
        <CalendarIcon className="size-3.5" />
        {text}
      </span>
    </SimpleTooltip>
  );
}
