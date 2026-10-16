import { useQuery } from "@tanstack/react-query";
import { BoxIcon } from "lucide-react";
import { type ReactNode, useState } from "react";

import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { List } from "@/components/layout/list";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { useNewProjectStore } from "@/components/projects/new-project-store";
import { groupProjectsByStatus } from "@/components/projects/project-groups";
import { useProjectStatuses } from "@/components/projects/project-pickers";
import { ProjectGroupHeader, ProjectRow } from "@/components/projects/project-row";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TListProjectsInput, projectsListQuery } from "@/lib/queries/projects";

export type TProjectListFilters = Pick<
  TListProjectsInput,
  | "team_id"
  | "status_type"
  | "status_id"
  | "lead_id"
  | "member_id"
  | "initiative_id"
  | "label_id"
  | "query"
  | "include_archived"
  | "include_trashed"
  | "order_by"
>;

type TProps = {
  workspaceId: string;
  filters: TProjectListFilters;
  emptyState?: ReactNode;
};

export function useProjectList(workspaceId: string, filters: TProjectListFilters) {
  const projectsQuery = useQuery(projectsListQuery({ workspace_id: workspaceId, ...filters }));
  const { statuses, isPending: statusesPending } = useProjectStatuses();
  const projects = projectsQuery.data?.projects ?? [];
  return {
    groups: groupProjectsByStatus(statuses, projects),
    projects,
    isPending: projectsQuery.isPending || statusesPending,
    error: projectsQuery.error,
    refetch: () => void projectsQuery.refetch(),
  };
}

export function ProjectList({ workspaceId, filters, emptyState }: TProps) {
  const { workspace, canEdit } = useWorkspace();
  const openNewProject = useNewProjectStore((state) => state.open);
  const { groups, projects, isPending, error, refetch } = useProjectList(workspaceId, filters);
  const [collapsed, setCollapsed] = useState<string[]>([]);

  if (isPending) return <ListSkeleton />;
  if (error) return <ErrorCard message={error.message} onRetry={refetch} />;
  if (projects.length === 0) {
    if (emptyState !== undefined) return emptyState;
    return (
      <EmptyState
        icon={BoxIcon}
        title="No projects yet"
        description="Projects group issues toward a goal."
      />
    );
  }

  const toggle = (statusId: string) =>
    setCollapsed((current) =>
      current.includes(statusId) ? current.filter((id) => id !== statusId) : [...current, statusId],
    );

  return (
    <List>
      {groups
        .filter((group) => group.projects.length > 0)
        .map((group) => {
          const isCollapsed = collapsed.includes(group.status.id);
          return (
            <div key={group.status.id} role="presentation" className="contents">
              <ProjectGroupHeader
                status={group.status}
                count={group.projects.length}
                isCollapsed={isCollapsed}
                onToggle={() => toggle(group.status.id)}
                onCreate={
                  canEdit
                    ? () =>
                        openNewProject({
                          statusId: group.status.id,
                          teamId: filters.team_id,
                          initiativeId: filters.initiative_id,
                        })
                    : undefined
                }
              />
              {!isCollapsed &&
                group.projects.map((project) => (
                  <ProjectRow key={project.id} workspaceSlug={workspace.slug} project={project} />
                ))}
            </div>
          );
        })}
    </List>
  );
}
