import { useQuery } from "@tanstack/react-query";
import { BoxIcon, PlusIcon } from "lucide-react";

import { PropertyPicker, type TPickerOption } from "@/components/issues/property-picker";
import { EmptyState } from "@/components/layout/empty-state";
import { SectionTitle } from "@/components/layout/section-title";
import { ProjectList } from "@/components/projects/project-list";
import { ProjectIcon } from "@/components/shared/project-icon";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  type TInitiative,
  useAddInitiativeProjectMutation,
  useRemoveInitiativeProjectMutation,
} from "@/lib/queries/initiatives";
import { projectsListQuery } from "@/lib/queries/projects";

export function InitiativeProjects({ initiative }: { initiative: TInitiative }) {
  const { workspace, canEdit } = useWorkspace();
  return (
    <section aria-label="Projects" className="flex w-full flex-col gap-4">
      <SectionTitle
        title="Projects"
        actions={canEdit && <ProjectsPicker initiative={initiative} />}
      />
      <ProjectList
        workspaceId={workspace.id}
        filters={{ initiative_id: initiative.id }}
        emptyState={
          <EmptyState
            icon={BoxIcon}
            title="No projects yet"
            description="Add the projects that move this initiative forward."
          />
        }
      />
    </section>
  );
}

function ProjectsPicker({ initiative }: { initiative: TInitiative }) {
  const { workspace } = useWorkspace();
  const { data, isPending } = useQuery(projectsListQuery({ workspace_id: workspace.id }));
  const { mutate: addProject } = useAddInitiativeProjectMutation();
  const { mutate: removeProject } = useRemoveInitiativeProjectMutation();
  const current = initiative.projects.map((project) => project.id);
  const options: TPickerOption[] = (data?.projects ?? []).map((project) => ({
    value: project.id,
    label: project.name,
    icon: <ProjectIcon icon={project.icon} color={project.color} />,
  }));

  const change = (next: string[]) => {
    const input = { initiativeId: initiative.id };
    next
      .filter((id) => !current.includes(id))
      .forEach((projectId) => addProject({ ...input, projectId }));
    current
      .filter((id) => !next.includes(id))
      .forEach((projectId) => removeProject({ ...input, projectId }));
  };

  return (
    <PropertyPicker
      label="Projects in this initiative"
      searchPlaceholder="Add or remove projects…"
      emptyText="No projects"
      multiple
      align="end"
      options={options}
      value={current}
      isPending={isPending}
      onChange={change}
      trigger={
        <Button variant="ghost" size="sm">
          <PlusIcon className="size-4" />
          Add project
        </Button>
      }
    />
  );
}
