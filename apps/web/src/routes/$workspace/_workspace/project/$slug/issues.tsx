import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";

import type { ListIssuesFilters } from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";
import {
  ProjectActions,
  projectBreadcrumb,
  ProjectTabs,
} from "@/components/projects/project-header";
import { ProjectLoader } from "@/components/projects/project-page";
import { ProjectIcon } from "@/components/shared/project-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import type { TProject } from "@/lib/queries/projects";

export const Route = createFileRoute("/$workspace/_workspace/project/$slug/issues")({
  component: ProjectSlugIssuesPage,
});

function ProjectSlugIssuesPage() {
  const { slug } = Route.useParams();
  return (
    <ProjectLoader slug={slug}>{(project) => <ProjectIssues project={project} />}</ProjectLoader>
  );
}

function ProjectIssues({ project }: { project: TProject }) {
  const { workspace } = useWorkspace();
  const [filters, setFilters] = useState<ListIssuesFilters>({});
  return (
    <IssuesView
      title={project.name}
      icon={<ProjectIcon icon={project.icon} color={project.color} size="lg" />}
      breadcrumb={projectBreadcrumb(workspace.slug, project)}
      storageKey={`project:${project.id}`}
      filters={filters}
      onFiltersChange={setFilters}
      fixedFilters={{ project_ids: [project.id] }}
      hiddenFilterFields={["project"]}
      tabs={
        <div className="flex w-full items-center gap-2">
          <div className="min-w-0 flex-1">
            <ProjectTabs project={project} tab="issues" />
          </div>
          <ProjectActions project={project} />
        </div>
      }
    />
  );
}
