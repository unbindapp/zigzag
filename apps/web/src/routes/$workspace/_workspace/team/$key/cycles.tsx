import { createFileRoute } from "@tanstack/react-router";
import { RefreshCwIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/cycles")({
  component: TeamKeyCyclesPage,
});

function TeamKeyCyclesPage() {
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Cycles"
        icon={<RefreshCwIcon />}
        breadcrumb={team ? [{ label: team.name }, { label: "Cycles" }] : undefined}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
