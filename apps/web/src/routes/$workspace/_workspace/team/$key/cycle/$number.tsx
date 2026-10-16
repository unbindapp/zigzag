import { createFileRoute } from "@tanstack/react-router";
import { RefreshCwIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/cycle/$number")({
  component: TeamCyclePage,
});

function TeamCyclePage() {
  const { workspace, key, number } = Route.useParams();
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title={`Cycle ${number}`}
        icon={<RefreshCwIcon />}
        breadcrumb={[
          { label: team?.name ?? key },
          {
            label: "Cycles",
            link: { to: "/$workspace/team/$key/cycles", params: { workspace, key } },
          },
          { label: `Cycle ${number}` },
        ]}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
