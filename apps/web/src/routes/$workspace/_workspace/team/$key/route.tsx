import { createFileRoute, Outlet } from "@tanstack/react-router";
import { UsersIcon } from "lucide-react";

import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { PageContainer } from "@/components/layout/page-container";
import { useTeams } from "@/components/workspace/teams-provider";

export const Route = createFileRoute("/$workspace/_workspace/team/$key")({
  component: TeamLayout,
});

function TeamLayout() {
  const { key } = Route.useParams();
  const { teamByKey, isPending, error, refetch } = useTeams();
  if (isPending) {
    return (
      <PageContainer>
        <ListSkeleton />
      </PageContainer>
    );
  }
  if (error) {
    return (
      <PageContainer>
        <ErrorCard message={error.message} onRetry={refetch} />
      </PageContainer>
    );
  }
  if (!teamByKey(key)) {
    return (
      <PageContainer>
        <EmptyState
          icon={UsersIcon}
          title="Team not found"
          description={`There is no team with the key ${key.toUpperCase()} in this workspace.`}
        />
      </PageContainer>
    );
  }
  return <Outlet />;
}
