import { createFileRoute } from "@tanstack/react-router";

import { ProjectsView } from "@/components/projects/projects-view";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/projects")({
  component: TeamKeyProjectsPage,
});

function TeamKeyProjectsPage() {
  const team = useRequiredTeamFromRoute();
  return (
    <ProjectsView
      title="Projects"
      breadcrumb={[{ label: team.name }, { label: "Projects" }]}
      storageKey={`team:${team.id}`}
      filters={{ team_id: team.id }}
    />
  );
}
