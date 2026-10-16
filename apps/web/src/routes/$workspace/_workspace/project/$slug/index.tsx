import { createFileRoute } from "@tanstack/react-router";

import { PageContainer } from "@/components/layout/page-container";
import { ProjectOverview } from "@/components/projects/project-overview";
import { ProjectLoader } from "@/components/projects/project-page";

export const Route = createFileRoute("/$workspace/_workspace/project/$slug/")({
  component: ProjectSlugPage,
});

function ProjectSlugPage() {
  const { slug } = Route.useParams();
  return (
    <ProjectLoader slug={slug}>
      {(project) => (
        <PageContainer>
          <ProjectOverview project={project} />
        </PageContainer>
      )}
    </ProjectLoader>
  );
}
