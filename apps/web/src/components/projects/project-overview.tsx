import { ProjectDocuments } from "@/components/documents/project-documents";
import { LifecycleBanner } from "@/components/documents/lifecycle-banner";
import { Breadcrumb } from "@/components/layout/breadcrumb";
import { AutosaveMarkdown } from "@/components/projects/autosave-markdown";
import { InlineText } from "@/components/projects/inline-text";
import { MilestonesSection } from "@/components/projects/milestones-section";
import {
  ProjectActions,
  projectBreadcrumb,
  ProjectTabs,
} from "@/components/projects/project-header";
import { ProjectProperties } from "@/components/projects/project-properties";
import { ProjectUpdatesSection } from "@/components/projects/project-updates-section";
import { IconPicker } from "@/components/shared/icon-picker";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  type TProject,
  projectNameMaxLength,
  useProjectLifecycleMutation,
  useUpdateProjectMutation,
} from "@/lib/queries/projects";

export function ProjectOverview({ project }: { project: TProject }) {
  const { workspace, canEdit } = useWorkspace();
  const { mutate: updateProject, mutateAsync: updateProjectAsync } = useUpdateProjectMutation();
  const { mutate: changeLifecycle, isPending: isRestoring } = useProjectLifecycleMutation();
  const lifecycleState = project.trashed_at ? "trashed" : project.archived_at ? "archived" : null;

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex w-full flex-col gap-2">
        <div className="flex w-full min-w-0 items-center gap-2">
          <div className="min-w-0 flex-1">
            <Breadcrumb items={projectBreadcrumb(workspace.slug, project)} />
          </div>
          <ProjectActions project={project} />
        </div>
        <ProjectTabs project={project} tab="overview" />
      </div>
      <div className="flex w-full flex-col gap-6 lg:flex-row">
        <div className="flex min-w-0 flex-1 flex-col gap-6">
          <LifecycleBanner
            entity="project"
            state={lifecycleState}
            isPending={isRestoring}
            onRestore={
              canEdit
                ? () =>
                    changeLifecycle({
                      projectId: project.id,
                      action: lifecycleState === "trashed" ? "restore" : "unarchive",
                    })
                : undefined
            }
          />
          <div className="flex flex-col gap-2">
            <div className="flex min-w-0 items-center gap-2">
              <IconPicker
                value={project.icon || null}
                color={project.color || null}
                disabled={!canEdit}
                label="Project icon"
                onChange={(icon) => updateProject({ project_id: project.id, icon: icon ?? "" })}
              />
              <InlineText
                label="Project name"
                value={project.name}
                maxLength={projectNameMaxLength}
                readOnly={!canEdit}
                onSave={(name) => updateProject({ project_id: project.id, name })}
              />
            </div>
            <InlineText
              label="Project summary"
              variant="summary"
              value={project.summary}
              maxLength={255}
              allowEmpty
              readOnly={!canEdit}
              placeholder={canEdit ? "Add a short summary…" : undefined}
              onSave={(summary) => updateProject({ project_id: project.id, summary })}
            />
            <AutosaveMarkdown
              key={project.id}
              label="Description"
              placeholder="Add a description…"
              initialValue={project.description}
              readOnly={!canEdit}
              onSave={(description) => updateProjectAsync({ project_id: project.id, description })}
            />
          </div>
          <MilestonesSection project={project} />
          <ProjectUpdatesSection project={project} />
          <ProjectDocuments workspaceId={workspace.id} projectId={project.id} />
        </div>
        <aside aria-label="Properties" className="w-full shrink-0 lg:w-72">
          <div className="rounded-lg border p-4 lg:sticky lg:top-0">
            <ProjectProperties project={project} />
          </div>
        </aside>
      </div>
    </div>
  );
}
