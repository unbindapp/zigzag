import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { Breadcrumb } from "@/components/layout/breadcrumb";
import { FavoriteButton } from "@/components/favorites/favorite-button";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import {
  ArchiveAction,
  CopyLinkAction,
  DeleteMenuAction,
  HeaderActions,
  TrashAction,
} from "@/components/shared/detail-actions";
import { Tabs, TabsList, TabsTab } from "@/components/ui/tabs";
import { toast } from "@/components/ui/toast";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  type TProject,
  type TProjectLifecycleAction,
  useProjectLifecycleMutation,
} from "@/lib/queries/projects";

export type TProjectTab = "overview" | "issues";

const lifecycleToasts: Record<Exclude<TProjectLifecycleAction, "delete">, string> = {
  archive: "Project archived",
  unarchive: "Project unarchived",
  trash: "Project moved to trash",
  restore: "Project restored",
};

export function ProjectActions({ project }: { project: TProject }) {
  const { workspace, canEdit, isAdmin } = useWorkspace();
  const navigate = useNavigate();
  const { mutateAsync: change, isPending } = useProjectLifecycleMutation();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);

  const run = async (action: Exclude<TProjectLifecycleAction, "delete">) => {
    await change({ projectId: project.id, action });
    toast.add({ type: "success", title: lifecycleToasts[action] });
  };

  return (
    <HeaderActions>
      <FavoriteButton kind="project" entityId={project.id} />
      <CopyLinkAction url={window.location.href.replace(/\/issues$/, "")} />
      {canEdit && (
        <ArchiveAction
          isOn={Boolean(project.archived_at)}
          isPending={isPending}
          onToggle={() =>
            void run(project.archived_at ? "unarchive" : "archive").catch(() => undefined)
          }
        />
      )}
      {canEdit && (
        <TrashAction
          isOn={Boolean(project.trashed_at)}
          isPending={isPending}
          onToggle={() => void run(project.trashed_at ? "restore" : "trash").catch(() => undefined)}
        />
      )}
      {isAdmin && (
        <>
          <DeleteMenuAction label="Delete project" onSelect={() => setIsDeleteOpen(true)} />
          <ConfirmDialog
            open={isDeleteOpen}
            onOpenChange={setIsDeleteOpen}
            title={`Delete ${project.name}`}
            description="Milestones, updates, links and documents of this project are deleted. Its issues stay and lose the project."
            confirmLabel="Delete project"
            onConfirm={async () => {
              await change({ projectId: project.id, action: "delete" });
              toast.add({ type: "success", title: "Project deleted" });
              await navigate({ to: "/$workspace/projects", params: { workspace: workspace.slug } });
            }}
          />
        </>
      )}
    </HeaderActions>
  );
}

export function ProjectTabs({ project, tab }: { project: TProject; tab: TProjectTab }) {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const params = { workspace: workspace.slug, slug: project.slug };
  return (
    <Tabs
      value={tab}
      onValueChange={(value) =>
        void navigate(
          value === "issues"
            ? { to: "/$workspace/project/$slug/issues", params }
            : { to: "/$workspace/project/$slug", params },
        )
      }
    >
      <TabsList>
        <TabsTab value="overview">Overview</TabsTab>
        <TabsTab value="issues">
          Issues
          <span className="text-muted-foreground text-xs font-normal">
            {project.issue_counts.total}
          </span>
        </TabsTab>
      </TabsList>
    </Tabs>
  );
}

export function ProjectHeaderBar({ project, tab }: { project: TProject; tab: TProjectTab }) {
  const { workspace } = useWorkspace();
  return (
    <div className="flex w-full flex-col gap-2">
      <div className="flex w-full min-w-0 items-center gap-2">
        <div className="min-w-0 flex-1">
          <Breadcrumb items={projectBreadcrumb(workspace.slug, project)} />
        </div>
        <ProjectActions project={project} />
      </div>
      <ProjectTabs project={project} tab={tab} />
    </div>
  );
}

export function projectBreadcrumb(workspaceSlug: string, project: TProject) {
  return [
    {
      label: "Projects",
      link: { to: "/$workspace/projects" as const, params: { workspace: workspaceSlug } },
    },
    { label: project.name },
  ];
}
