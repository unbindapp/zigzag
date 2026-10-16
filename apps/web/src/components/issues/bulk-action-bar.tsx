import { ArchiveIcon, Trash2Icon, XIcon } from "lucide-react";

import { patchToOptimistic, type TIssuePatch } from "@/components/issues/issue-draft";
import {
  AssigneePicker,
  CyclePicker,
  LabelsPicker,
  PriorityPicker,
  ProjectPicker,
  StatePicker,
} from "@/components/issues/pickers";
import {
  type TBulkUpdateIssuesInput,
  type TIssue,
  CLEAR_UUID,
  useBulkUpdateIssuesMutation,
} from "@/lib/queries/issues";
import { useIssueLifecycle } from "@/components/issues/use-issue-actions";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";

type TProps = {
  issues: TIssue[];
  onClear: () => void;
};

function shared<T>(issues: TIssue[], pick: (issue: TIssue) => T, key: (value: T) => string) {
  const first = issues[0] ? pick(issues[0]) : undefined;
  if (first === undefined) return undefined;
  const firstKey = key(first);
  return issues.every((issue) => key(pick(issue)) === firstKey) ? first : undefined;
}

export function BulkActionBar({ issues, onClear }: TProps) {
  const { canEdit } = useWorkspace();
  const { mutate: bulkUpdate } = useBulkUpdateIssuesMutation();
  const lifecycle = useIssueLifecycle();
  if (issues.length === 0 || !canEdit) return null;

  const issueIds = issues.map((issue) => issue.id);
  const teamId = shared(
    issues,
    (issue) => issue.team_id,
    (id) => id,
  );
  const state = shared(
    issues,
    (issue) => issue.state,
    (value) => value.id,
  );
  const priority = shared(issues, (issue) => issue.priority, String);
  const assignee = shared(
    issues,
    (issue) => issue.assignee ?? null,
    (user) => user?.id ?? "",
  );
  const project = shared(
    issues,
    (issue) => issue.project ?? null,
    (item) => item?.id ?? "",
  );
  const cycle = shared(
    issues,
    (issue) => issue.cycle ?? null,
    (item) => item?.id ?? "",
  );
  const commonLabels = issues[0].labels.filter((label) =>
    issues.every((issue) => issue.labels.some((item) => item.id === label.id)),
  );

  const apply = (input: Omit<TBulkUpdateIssuesInput, "issue_ids">, patch?: TIssuePatch) =>
    bulkUpdate({
      input: { issue_ids: issueIds, ...input },
      optimistic: patch ? patchToOptimistic(patch) : undefined,
    });

  return (
    <div className="pointer-events-none fixed inset-x-0 bottom-4 z-40 flex justify-center px-4">
      <div
        role="toolbar"
        aria-label="Bulk actions"
        className="bg-popover pointer-events-auto flex max-w-full flex-wrap items-center gap-2 rounded-lg border p-2 shadow-lg"
      >
        <span className="px-1 text-sm font-medium whitespace-nowrap">{issues.length} selected</span>
        <Button variant="ghost" size="icon" aria-label="Clear selection" onClick={onClear}>
          <XIcon className="size-4" />
        </Button>
        {teamId && (
          <StatePicker
            appearance="compact"
            teamId={teamId}
            state={state}
            onChange={(next) => apply({ state_id: next.id }, { state: next })}
          />
        )}
        <PriorityPicker
          appearance="compact"
          priority={priority ?? 0}
          onChange={(next) => apply({ priority: next }, { priority: next })}
        />
        <AssigneePicker
          appearance="compact"
          assignee={assignee}
          onChange={(next) => apply({ assignee_id: next?.id ?? CLEAR_UUID }, { assignee: next })}
        />
        <LabelsPicker
          appearance="compact"
          teamId={teamId}
          labels={commonLabels}
          onChange={(next) => {
            const nextIds = next.map((label) => label.id);
            const commonIds = commonLabels.map((label) => label.id);
            apply({
              add_label_ids: nextIds.filter((id) => !commonIds.includes(id)),
              remove_label_ids: commonIds.filter((id) => !nextIds.includes(id)),
            });
          }}
        />
        <ProjectPicker
          appearance="compact"
          teamId={teamId}
          project={project}
          onChange={(next) => apply({ project_id: next?.id ?? CLEAR_UUID }, { project: next })}
        />
        {teamId && (
          <CyclePicker
            appearance="compact"
            teamId={teamId}
            cycle={cycle}
            onChange={(next) => apply({ cycle_id: next?.id ?? CLEAR_UUID }, { cycle: next })}
          />
        )}
        <Button
          variant="outline"
          size="sm"
          isPending={lifecycle.isPending}
          onClick={() => void lifecycle.run(issueIds, "archive").then(onClear, () => undefined)}
        >
          <ArchiveIcon className="size-4" />
          Archive
        </Button>
        <Button
          variant="outline"
          size="sm"
          isPending={lifecycle.isPending}
          onClick={() => void lifecycle.run(issueIds, "trash").then(onClear, () => undefined)}
        >
          <Trash2Icon className="size-4" />
          Trash
        </Button>
      </div>
    </div>
  );
}
