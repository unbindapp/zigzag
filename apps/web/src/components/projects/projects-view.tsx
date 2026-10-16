import { BoxIcon, LayoutGridIcon, ListIcon, PlusIcon } from "lucide-react";
import { useLocalStorage } from "usehooks-ts";

import type { TBreadcrumbItem } from "@/components/layout/breadcrumb";
import { EmptyState } from "@/components/layout/empty-state";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { PageToolbar } from "@/components/layout/page-toolbar";
import { NewProjectDialog } from "@/components/projects/new-project-dialog";
import { useNewProjectStore } from "@/components/projects/new-project-store";
import { ProjectBoard } from "@/components/projects/project-board";
import { ProjectList, type TProjectListFilters } from "@/components/projects/project-list";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";

type TLayout = "list" | "board";

type TProps = {
  title: string;
  breadcrumb?: TBreadcrumbItem[];
  storageKey: string;
  filters: TProjectListFilters;
};

export function ProjectsView({ title, breadcrumb, storageKey, filters }: TProps) {
  const { workspace, canEdit } = useWorkspace();
  const openNewProject = useNewProjectStore((state) => state.open);
  const [layout, setLayout] = useLocalStorage<TLayout>(`projects-layout:${storageKey}`, "list");
  const create = () => openNewProject({ teamId: filters.team_id });
  const emptyState = (
    <EmptyState
      icon={BoxIcon}
      title="No projects yet"
      description="Projects group issues from one or more teams toward a goal."
      action={
        canEdit && (
          <Button size="sm" variant="outline" onClick={create}>
            <PlusIcon className="size-4" />
            New project
          </Button>
        )
      }
    />
  );
  const listProps = { workspaceId: workspace.id, filters, emptyState };

  return (
    <PageContainer>
      <PageHeader
        title={title}
        icon={<BoxIcon />}
        breadcrumb={breadcrumb}
        actions={
          canEdit && (
            <Button onClick={create}>
              <PlusIcon className="size-4" />
              New project
            </Button>
          )
        }
        toolbar={
          <PageToolbar
            actions={
              <div role="radiogroup" aria-label="Layout" className="flex items-center gap-2">
                <Button
                  variant={layout === "list" ? "secondary" : "ghost"}
                  size="icon"
                  role="radio"
                  aria-checked={layout === "list"}
                  aria-label="List layout"
                  onClick={() => setLayout("list")}
                >
                  <ListIcon className="size-4" />
                </Button>
                <Button
                  variant={layout === "board" ? "secondary" : "ghost"}
                  size="icon"
                  role="radio"
                  aria-checked={layout === "board"}
                  aria-label="Board layout"
                  onClick={() => setLayout("board")}
                >
                  <LayoutGridIcon className="size-4" />
                </Button>
              </div>
            }
          />
        }
      />
      {layout === "board" ? <ProjectBoard {...listProps} /> : <ProjectList {...listProps} />}
      <NewProjectDialog />
    </PageContainer>
  );
}
