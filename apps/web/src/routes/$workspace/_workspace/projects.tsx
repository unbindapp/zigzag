import { createFileRoute } from "@tanstack/react-router";

import { ProjectsView } from "@/components/projects/projects-view";

export const Route = createFileRoute("/$workspace/_workspace/projects")({
  component: ProjectsPage,
});

function ProjectsPage() {
  return <ProjectsView title="Projects" storageKey="workspace" filters={{}} />;
}
