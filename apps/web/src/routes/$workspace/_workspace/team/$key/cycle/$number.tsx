import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { RefreshCwIcon } from "lucide-react";

import { CycleDetail } from "@/components/cycles/cycle-detail";
import { CycleMenu } from "@/components/cycles/cycle-menu";
import { cycleStatus } from "@/components/cycles/cycle-status";
import { FavoriteButton } from "@/components/favorites/favorite-button";
import { IssuePeekHost } from "@/components/issues/issue-peek";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { CycleIcon } from "@/components/shared/cycle-icon";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { cycleByNumberQuery, cycleName } from "@/lib/queries/cycles";
import { isApiError } from "@/lib/server/errors";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/cycle/$number")({
  component: TeamCyclePage,
});

function TeamCyclePage() {
  const { workspace: workspaceSlug, key, number } = Route.useParams();
  const { workspace, isAdmin } = useWorkspace();
  const team = useRequiredTeamFromRoute();
  const cycleNumber = Number(number);
  const isValidNumber = Number.isInteger(cycleNumber) && cycleNumber > 0;
  const { data, isPending, error, refetch } = useQuery({
    ...cycleByNumberQuery(team.id, cycleNumber),
    enabled: isValidNumber,
  });
  const cycle = data?.cycle;
  const title = cycle ? cycleName(cycle) : `Cycle ${number}`;

  const body = () => {
    if (!isValidNumber || isApiError(error, 404)) {
      return (
        <EmptyState
          icon={RefreshCwIcon}
          title="Cycle not found"
          description={`${team.name} has no cycle ${number}.`}
        />
      );
    }
    if (isPending) return <ListSkeleton />;
    if (error) return <ErrorCard message={error.message} onRetry={() => void refetch()} />;
    return <CycleDetail cycle={data.cycle} team={team} workspaceId={workspace.id} />;
  };

  return (
    <IssuePeekHost>
      <PageContainer>
        <PageHeader
          title={title}
          icon={
            cycle ? (
              <CycleIcon
                progress={cycle.progress.percent / 100}
                status={cycleStatus(cycle)}
                className="size-5"
              />
            ) : (
              <RefreshCwIcon />
            )
          }
          breadcrumb={[
            { label: team.name },
            {
              label: "Cycles",
              link: {
                to: "/$workspace/team/$key/cycles",
                params: { workspace: workspaceSlug, key },
              },
            },
            { label: title },
          ]}
          actions={
            cycle && (
              <>
                <FavoriteButton kind="cycle" entityId={cycle.id} />
                {isAdmin && (
                  <CycleMenu cycle={cycle} workspaceSlug={workspaceSlug} teamKey={team.key} />
                )}
              </>
            )
          }
        />
        {body()}
      </PageContainer>
    </IssuePeekHost>
  );
}
