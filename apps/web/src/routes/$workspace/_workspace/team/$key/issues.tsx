import { createFileRoute } from "@tanstack/react-router";
import { CircleDotIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/issues")({
  component: TeamKeyIssuesPage,
});

function TeamKeyIssuesPage() {
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Issues"
        icon={<CircleDotIcon />}
        breadcrumb={team ? [{ label: team.name }, { label: "Issues" }] : undefined}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
