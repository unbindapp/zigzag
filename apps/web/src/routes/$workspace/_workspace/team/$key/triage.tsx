import { createFileRoute } from "@tanstack/react-router";
import { TriangleIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/triage")({
  component: TeamKeyTriagePage,
});

function TeamKeyTriagePage() {
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Triage"
        icon={<TriangleIcon />}
        breadcrumb={team ? [{ label: team.name }, { label: "Triage" }] : undefined}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
