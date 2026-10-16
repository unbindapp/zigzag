import { LabelsPicker, PriorityPicker } from "@/components/issues/pickers";
import { PropertyRow } from "@/components/layout/property-row";
import { HealthBadge } from "@/components/projects/health-badge";
import { ProjectLinks } from "@/components/projects/project-links";
import {
  LeadPicker,
  MembersPicker,
  ProjectDatePicker,
  ProjectStatusPicker,
  TeamsPicker,
} from "@/components/projects/project-pickers";
import { ProgressBar } from "@/components/projects/progress-bar";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  type TProject,
  type TUpdateProjectInput,
  useUpdateProjectMutation,
} from "@/lib/queries/projects";

const NIL_UUID = "00000000-0000-0000-0000-000000000000";

export function ProjectProperties({ project }: { project: TProject }) {
  const { canEdit } = useWorkspace();
  const { mutate: updateProject } = useUpdateProjectMutation();
  const update = (input: Omit<TUpdateProjectInput, "project_id">) =>
    updateProject({ project_id: project.id, ...input });
  const disabled = !canEdit;

  return (
    <div className="flex flex-col gap-1">
      <PropertyRow label="Status">
        <ProjectStatusPicker
          status={project.status}
          disabled={disabled}
          onChange={(status) => update({ status_id: status.id })}
        />
      </PropertyRow>
      <PropertyRow label="Health">
        <HealthBadge health={project.health} />
      </PropertyRow>
      <PropertyRow label="Priority">
        <PriorityPicker
          priority={project.priority}
          disabled={disabled}
          onChange={(priority) => update({ priority })}
        />
      </PropertyRow>
      <PropertyRow label="Lead">
        <LeadPicker
          lead={project.lead}
          disabled={disabled}
          onChange={(user) => update({ lead_id: user?.id ?? NIL_UUID })}
        />
      </PropertyRow>
      <PropertyRow label="Members">
        <MembersPicker
          members={project.members}
          disabled={disabled}
          onChange={(ids) => update({ member_ids: ids })}
        />
      </PropertyRow>
      <PropertyRow label="Teams">
        <TeamsPicker
          teams={project.teams}
          disabled={disabled}
          onChange={(ids) => update({ team_ids: ids })}
        />
      </PropertyRow>
      <PropertyRow label="Start date">
        <ProjectDatePicker
          label="Start date"
          date={project.start_date}
          resolution={project.start_date_resolution}
          disabled={disabled}
          onChange={({ date, resolution }) =>
            update({ start_date: date ?? "", start_date_resolution: resolution })
          }
        />
      </PropertyRow>
      <PropertyRow label="Target date">
        <ProjectDatePicker
          label="Target date"
          date={project.target_date}
          resolution={project.target_date_resolution}
          disabled={disabled}
          onChange={({ date, resolution }) =>
            update({ target_date: date ?? "", target_date_resolution: resolution })
          }
        />
      </PropertyRow>
      <PropertyRow label="Labels">
        <LabelsPicker
          teamId={undefined}
          labels={project.labels}
          disabled={disabled}
          onChange={(labels) => update({ label_ids: labels.map((label) => label.id) })}
        />
      </PropertyRow>
      <PropertyRow label="Progress">
        <ProgressBar percent={project.progress.percent} label="Project progress" />
      </PropertyRow>
      <div className="flex flex-col gap-1.5 pt-2">
        <span className="text-muted-foreground text-sm">Links</span>
        <ProjectLinks project={project} canEdit={canEdit} />
      </div>
    </div>
  );
}
