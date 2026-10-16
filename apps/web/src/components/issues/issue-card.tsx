import type { CSSProperties, HTMLAttributes, MouseEvent, Ref } from "react";

import type { TDisplaySettings } from "@/components/issues/display";
import { DueDateText, ProjectChip } from "@/components/issues/issue-chips";
import { AssigneePicker, PriorityPicker, StatePicker } from "@/components/issues/pickers";
import type { TIssue } from "@/lib/queries/issues";
import { useIssuePatch } from "@/components/issues/use-issue-actions";
import { LabelChip } from "@/components/shared/label-chip";
import { cn } from "@/components/ui/utils";

type TProps = {
  issue: TIssue;
  properties: TDisplaySettings["properties"];
  isActive?: boolean;
  isDragging?: boolean;
  isOverlay?: boolean;
  canEdit: boolean;
  onClick?: (event: MouseEvent, issue: TIssue) => void;
  ref?: Ref<HTMLDivElement>;
  style?: CSSProperties;
  dragProps?: HTMLAttributes<HTMLDivElement>;
};

export function IssueCard({
  issue,
  properties,
  isActive,
  isDragging,
  isOverlay,
  canEdit,
  onClick,
  ref,
  style,
  dragProps,
}: TProps) {
  const patch = useIssuePatch();
  const show = (property: TDisplaySettings["properties"][number]) => properties.includes(property);
  return (
    <div
      ref={ref}
      style={style}
      {...dragProps}
      role="listitem"
      aria-label={`${issue.identifier} ${issue.title}`}
      data-active={isActive || undefined}
      data-dragging={isDragging || undefined}
      onClick={(event) => onClick?.(event, issue)}
      className={cn(
        "bg-card has-hover:hover:border-primary/4-10 data-active:border-primary/6-10 focus-visible:ring-primary/8-10 flex w-full cursor-default flex-col gap-2 rounded-lg border p-4 text-sm outline-hidden focus-visible:ring-1 data-dragging:opacity-40",
        isOverlay && "shadow-lg",
      )}
    >
      <div className="flex items-center gap-2">
        {show("identifier") && (
          <span className="text-muted-foreground min-w-0 flex-1 truncate font-mono text-xs">
            {issue.identifier}
          </span>
        )}
        {show("assignee") && (
          <AssigneePicker
            appearance="icon"
            disabled={!canEdit}
            assignee={issue.assignee}
            onChange={(assignee) => patch(issue, { assignee })}
          />
        )}
      </div>
      <div className="flex items-start gap-2">
        {show("state") && (
          <StatePicker
            appearance="icon"
            disabled={!canEdit}
            teamId={issue.team_id}
            state={issue.state}
            onChange={(state) => patch(issue, { state })}
          />
        )}
        <p className="min-w-0 flex-1 self-center">{issue.title}</p>
      </div>
      <div className="flex flex-wrap items-center gap-1.5">
        {show("priority") && (
          <PriorityPicker
            appearance="icon"
            disabled={!canEdit}
            priority={issue.priority}
            onChange={(priority) => patch(issue, { priority })}
          />
        )}
        {show("labels") &&
          issue.labels.map((label) => (
            <LabelChip key={label.id} name={label.name} color={label.color} />
          ))}
        {show("project") && issue.project && <ProjectChip project={issue.project} />}
        {show("due_date") && <DueDateText issue={issue} />}
      </div>
    </div>
  );
}
