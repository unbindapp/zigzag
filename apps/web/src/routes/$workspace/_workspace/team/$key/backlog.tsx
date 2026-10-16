import { createFileRoute } from "@tanstack/react-router";
import { ListTodoIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/backlog")({
  component: TeamKeyBacklogPage,
});

function TeamKeyBacklogPage() {
  const team = useTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Backlog"
        icon={<ListTodoIcon />}
        breadcrumb={team ? [{ label: team.name }, { label: "Backlog" }] : undefined}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
