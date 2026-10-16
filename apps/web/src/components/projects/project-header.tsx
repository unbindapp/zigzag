import { useNavigate } from "@tanstack/react-router";
import {
  ArchiveIcon,
  ArchiveRestoreIcon,
  LinkIcon,
  Trash2Icon,
  UndoIcon,
  XIcon,
} from "lucide-react";
import { type ReactNode, useState } from "react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTab } from "@/components/ui/tabs";
import { toast } from "@/components/ui/toast";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useCopyToClipboard } from "@/lib/hooks/use-copy";
import {
  type TProject,
  type TProjectLifecycleAction,
  useProjectLifecycleMutation,
} from "@/lib/queries/projects";

export type TProjectTab = "overview" | "issues";

function HeaderAction({
  label,
  onClick,
  isPending,
  children,
}: {
  label: string;
  onClick?: () => void;
  isPending?: boolean;
  children: ReactNode;
}) {
  return (
    <SimpleTooltip content={label}>
      <Button
        variant="ghost"
        size="icon"
        aria-label={label}
        isPending={isPending}
        onClick={onClick}
      >
        {children}
      </Button>
    </SimpleTooltip>
  );
}

const lifecycleToasts: Record<Exclude<TProjectLifecycleAction, "delete">, string> = {
  archive: "Project archived",
  unarchive: "Project unarchived",
  trash: "Project moved to trash",
  restore: "Project restored",
};

export function ProjectActions({ project }: { project: TProject }) {
  const { workspace, canEdit, isAdmin } = useWorkspace();
  const navigate = useNavigate();
  const { copyToClipboard } = useCopyToClipboard();
  const { mutateAsync: change, isPending } = useProjectLifecycleMutation();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);

  const run = async (action: Exclude<TProjectLifecycleAction, "delete">) => {
    await change({ projectId: project.id, action });
    toast.add({ type: "success", title: lifecycleToasts[action] });
  };

  return (
    <div className="flex shrink-0 items-center gap-1">
      <FavoriteButton kind="project" entityId={project.id} />
      <HeaderAction
        label="Copy link"
        onClick={() => {
          copyToClipboard(window.location.href.replace(/\/issues$/, ""));
          toast.add({ type: "success", title: "Link copied" });
        }}
      >
        <LinkIcon className="size-4" />
      </HeaderAction>
      {canEdit && (
        <HeaderAction
          label={project.archived_at ? "Unarchive" : "Archive"}
          isPending={isPending}
          onClick={() =>
            void run(project.archived_at ? "unarchive" : "archive").catch(() => undefined)
          }
        >
          {project.archived_at ? (
            <ArchiveRestoreIcon className="size-4" />
          ) : (
            <ArchiveIcon className="size-4" />
          )}
        </HeaderAction>
      )}
      {canEdit && (
        <HeaderAction
          label={project.trashed_at ? "Restore from trash" : "Move to trash"}
          isPending={isPending}
          onClick={() => void run(project.trashed_at ? "restore" : "trash").catch(() => undefined)}
        >
          {project.trashed_at ? <UndoIcon className="size-4" /> : <Trash2Icon className="size-4" />}
        </HeaderAction>
      )}
      {isAdmin && (
        <>
          <HeaderAction label="Delete project" onClick={() => setIsDeleteOpen(true)}>
            <XIcon className="size-4" />
          </HeaderAction>
          <ConfirmDialog
            open={isDeleteOpen}
            onOpenChange={setIsDeleteOpen}
            title={`Delete ${project.name}`}
            description="Milestones, updates, links and documents of this project are deleted. Its issues stay and lose the project."
            confirmLabel="Delete project"
            confirmText={project.name}
            onConfirm={async () => {
              await change({ projectId: project.id, action: "delete" });
              toast.add({ type: "success", title: "Project deleted" });
              await navigate({ to: "/$workspace/projects", params: { workspace: workspace.slug } });
            }}
          />
        </>
      )}
    </div>
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

export function projectBreadcrumb(workspaceSlug: string, project: TProject) {
  return [
    {
      label: "Projects",
      link: { to: "/$workspace/projects" as const, params: { workspace: workspaceSlug } },
    },
    { label: project.name },
  ];
}
