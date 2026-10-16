import { createFileRoute } from "@tanstack/react-router";

import { ProjectsView } from "@/components/projects/projects-view";
import { teamCrumb } from "@/components/workspace/team-crumb";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/projects")({
  component: TeamKeyProjectsPage,
});

function TeamKeyProjectsPage() {
  const team = useRequiredTeamFromRoute();
  const { workspace } = Route.useParams();
  return (
    <ProjectsView
      title="Projects"
      breadcrumb={[teamCrumb(team, workspace), { label: "Projects" }]}
      storageKey={`team:${team.id}`}
      filters={{ team_id: team.id }}
    />
  );
}
