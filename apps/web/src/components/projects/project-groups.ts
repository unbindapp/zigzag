import type { TProjectListItem, TProjectStatus } from "@/lib/queries/projects";

export type TProjectGroup = { status: TProjectStatus; projects: TProjectListItem[] };

export function groupProjectsByStatus(statuses: TProjectStatus[], projects: TProjectListItem[]) {
  const known = new Set(statuses.map((status) => status.id));
  const extra = projects
    .map((project) => project.status)
    .filter(
      (status, index, all) =>
        !known.has(status.id) && all.findIndex((s) => s.id === status.id) === index,
    );
  return [...statuses, ...extra].map((status) => ({
    status,
    projects: projects.filter((project) => project.status.id === status.id),
  }));
}
