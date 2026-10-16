import { createFileRoute } from "@tanstack/react-router";
import { TagIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/labels")({
  component: TeamKeyLabelsPage,
});

function TeamKeyLabelsPage() {
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Labels"
        icon={<TagIcon />}
        breadcrumb={team ? [{ label: team.name }, { label: "Labels" }] : undefined}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
