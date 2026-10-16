import type { MouseEvent as ReactMouseEvent } from "react";

import type { TDisplaySettings } from "@/components/issues/display";
import { DueDateText, ProjectChip } from "@/components/issues/issue-chips";
import { AssigneePicker, PriorityPicker, StatePicker } from "@/components/issues/pickers";
import type { TIssue } from "@/lib/queries/issues";
import { useIssuePatch } from "@/components/issues/use-issue-actions";
import { LabelChip } from "@/components/shared/label-chip";
import { RelativeTime } from "@/components/shared/relative-time";
import { EstimateBadge } from "@/components/shared/estimate-badge";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/components/ui/utils";
import { useTeams } from "@/components/workspace/teams-provider";
import type { TEstimateType } from "@/lib/helpers/estimates";

const MAX_ROW_LABELS = 3;

type TProps = {
  issue: TIssue;
  depth?: number;
  properties: TDisplaySettings["properties"];
  isSelected: boolean;
  isSelecting: boolean;
  isActive: boolean;
  canEdit: boolean;
  onClick: (event: ReactMouseEvent, issue: TIssue) => void;
  onToggleSelect: (issue: TIssue, extendRange: boolean) => void;
  onFocus: (issue: TIssue) => void;
};

export function IssueRow({
  issue,
  depth = 0,
  properties,
  isSelected,
  isSelecting,
  isActive,
  canEdit,
  onClick,
  onToggleSelect,
  onFocus,
}: TProps) {
  const patch = useIssuePatch();
  const { teams } = useTeams();
  const team = teams.find((item) => item.id === issue.team_id);
  const show = (property: TDisplaySettings["properties"][number]) => properties.includes(property);
  const labels = issue.labels.slice(0, MAX_ROW_LABELS);
  const hiddenLabels = issue.labels.length - labels.length;

  return (
    <div
      role="listitem"
      tabIndex={0}
      aria-selected={isSelected}
      aria-label={`${issue.identifier} ${issue.title}`}
      data-selected={isSelected || undefined}
      data-active={isActive || undefined}
      onClick={(event) => onClick(event, issue)}
      onKeyDown={(event) => {
        if (event.key !== "Enter" || event.target !== event.currentTarget) return;
        onClick(event as unknown as ReactMouseEvent, issue);
      }}
      onMouseEnter={() => onFocus(issue)}
      onFocus={() => onFocus(issue)}
      className="group/row has-hover:hover:bg-accent data-selected:bg-primary/1-10 data-active:bg-accent focus-visible:ring-primary/8-10 relative flex h-9 w-full cursor-default items-center gap-2 border-b pr-3 pl-6 text-sm outline-hidden focus-visible:ring-1 focus-visible:ring-inset"
    >
      <span
        className="absolute top-1/2 left-1.5 flex -translate-y-1/2 items-center"
        onClick={(event) => event.stopPropagation()}
      >
        <Checkbox
          aria-label={`Select ${issue.identifier}`}
          checked={isSelected}
          onCheckedChange={(_checked, details) =>
            onToggleSelect(issue, (details.event as MouseEvent | undefined)?.shiftKey ?? false)
          }
          className={cn(
            "opacity-0 group-has-hover/row:group-hover/row:opacity-100 focus-visible:opacity-100",
            (isSelected || isSelecting) && "opacity-100",
          )}
        />
      </span>
      {show("priority") && (
        <PriorityPicker
          appearance="icon"
          disabled={!canEdit}
          priority={issue.priority}
          onChange={(priority) => patch(issue, { priority })}
        />
      )}
      {show("identifier") && (
        <span className="text-muted-foreground w-16 shrink-0 truncate font-mono text-xs">
          {issue.identifier}
        </span>
      )}
      {show("state") && (
        <StatePicker
          appearance="icon"
          disabled={!canEdit}
          teamId={issue.team_id}
          state={issue.state}
          onChange={(state) => patch(issue, { state })}
        />
      )}
      <span
        data-nested={depth > 0 || undefined}
        className="min-w-0 flex-1 truncate data-nested:pl-4"
      >
        {issue.title}
      </span>
      <span className="hidden min-w-0 shrink items-center justify-end gap-1.5 overflow-hidden sm:flex">
        {show("labels") &&
          labels.map((label) => <LabelChip key={label.id} name={label.name} color={label.color} />)}
        {show("labels") && hiddenLabels > 0 && (
          <span className="text-muted-foreground text-xs">+{hiddenLabels}</span>
        )}
        {show("project") && issue.project && <ProjectChip project={issue.project} />}
        {show("cycle") && issue.cycle && (
          <span className="text-muted-foreground shrink-0 text-xs whitespace-nowrap">
            {issue.cycle.name || `Cycle ${issue.cycle.number}`}
          </span>
        )}
        {show("estimate") && team && (
          <EstimateBadge value={issue.estimate} type={team.estimate_type as TEstimateType} />
        )}
        {show("due_date") && <DueDateText issue={issue} />}
        {show("updated_at") && (
          <span className="text-muted-foreground shrink-0 text-xs">
            <RelativeTime value={issue.updated_at} />
          </span>
        )}
      </span>
      {show("assignee") && (
        <AssigneePicker
          appearance="icon"
          disabled={!canEdit}
          assignee={issue.assignee}
          onChange={(assignee) => patch(issue, { assignee })}
        />
      )}
    </div>
  );
}
