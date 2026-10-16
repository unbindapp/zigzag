import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";

import { filtersFromSearch, filtersToSearch, issueSearchSchema } from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";
import { ProjectHeaderBar } from "@/components/projects/project-header";
import { ProjectLoader } from "@/components/projects/project-page";
import { ProjectIcon } from "@/components/shared/project-icon";
import type { TProject } from "@/lib/queries/projects";

export const Route = createFileRoute("/$workspace/_workspace/project/$slug/issues")({
  validateSearch: zodValidator(issueSearchSchema),
  component: ProjectSlugIssuesPage,
});

function ProjectSlugIssuesPage() {
  const { slug } = Route.useParams();
  return (
    <ProjectLoader slug={slug}>{(project) => <ProjectIssues project={project} />}</ProjectLoader>
  );
}

function ProjectIssues({ project }: { project: TProject }) {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  return (
    <IssuesView
      title={project.name}
      icon={<ProjectIcon icon={project.icon} color={project.color} size="lg" />}
      header={<ProjectHeaderBar project={project} tab="issues" />}
      storageKey={`project:${project.id}`}
      filters={filtersFromSearch(search)}
      onFiltersChange={(filters) =>
        void navigate({ search: (prev) => ({ ...prev, ...filtersToSearch(filters) }) })
      }
      fixedFilters={{ project_ids: [project.id] }}
      hiddenFilterFields={["project"]}
    />
  );
}
