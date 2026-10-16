import { DiamondIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";

import { EmptyState } from "@/components/layout/empty-state";
import { List } from "@/components/layout/list";
import { SectionTitle } from "@/components/layout/section-title";
import { MilestoneDialog } from "@/components/projects/milestone-dialog";
import { ProgressBar } from "@/components/projects/progress-bar";
import { formatProjectDate } from "@/components/projects/project-dates";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TMilestone, type TProject, useDeleteMilestoneMutation } from "@/lib/queries/projects";

export function MilestonesSection({ project }: { project: TProject }) {
  const { canEdit } = useWorkspace();
  return (
    <section aria-label="Milestones" className="flex w-full flex-col gap-4">
      <SectionTitle
        title="Milestones"
        actions={
          canEdit && (
            <MilestoneDialog
              mode="create"
              projectId={project.id}
              trigger={
                <Button variant="ghost" size="sm">
                  <PlusIcon className="size-4" />
                  Add milestone
                </Button>
              }
            />
          )
        }
      />
      {project.milestones.length === 0 ? (
        <EmptyState
          icon={DiamondIcon}
          title="No milestones"
          description="Split the project into stages and track each one's progress."
        />
      ) : (
        <List>
          {project.milestones.map((milestone) => (
            <MilestoneRow key={milestone.id} milestone={milestone} canEdit={canEdit} />
          ))}
        </List>
      )}
    </section>
  );
}

function MilestoneRow({ milestone, canEdit }: { milestone: TMilestone; canEdit: boolean }) {
  const { mutateAsync: deleteMilestone } = useDeleteMilestoneMutation();
  return (
    <div
      role="listitem"
      className="flex h-9 w-full items-center gap-2 border-b px-3 text-sm last:border-b-0"
    >
      <DiamondIcon className="text-muted-foreground size-4 shrink-0" />
      <span className="min-w-0 flex-1 truncate">{milestone.name}</span>
      <span className="text-muted-foreground hidden shrink-0 text-xs sm:inline">
        {milestone.issue_count} issue{milestone.issue_count === 1 ? "" : "s"}
      </span>
      <span className="text-muted-foreground hidden w-24 shrink-0 truncate text-xs sm:block">
        {milestone.target_date ? formatProjectDate(milestone.target_date) : ""}
      </span>
      <span className="w-28 shrink-0">
        <ProgressBar percent={milestone.progress.percent} label={`${milestone.name} progress`} />
      </span>
      {canEdit && (
        <div className="flex shrink-0 items-center gap-1">
          <MilestoneDialog
            mode="edit"
            milestone={milestone}
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Edit ${milestone.name}`}>
                <PencilIcon className="size-4" />
              </Button>
            }
          />
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Delete ${milestone.name}`}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title={`Delete ${milestone.name}`}
            description="Its issues stay in the project without a milestone."
            confirmLabel="Delete milestone"
            onConfirm={() => deleteMilestone({ milestoneId: milestone.id })}
          />
        </div>
      )}
    </div>
  );
}
