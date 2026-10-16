import { createFileRoute } from "@tanstack/react-router";
import { BoxIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/projects")({
  component: TeamKeyProjectsPage,
});

function TeamKeyProjectsPage() {
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Projects"
        icon={<BoxIcon />}
        breadcrumb={team ? [{ label: team.name }, { label: "Projects" }] : undefined}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
