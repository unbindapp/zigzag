import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { PlusIcon, RefreshCwIcon } from "lucide-react";

import { CycleDialog } from "@/components/cycles/cycle-dialog";
import { CycleList } from "@/components/cycles/cycle-list";
import { EmptyState } from "@/components/layout/empty-state";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { Button, LinkButton } from "@/components/ui/button";
import { teamCrumb } from "@/components/workspace/team-crumb";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { cyclesListQuery } from "@/lib/queries/cycles";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/cycles")({
  component: TeamKeyCyclesPage,
});

function TeamKeyCyclesPage() {
  const { workspace, isAdmin } = useWorkspace();
  const team = useRequiredTeamFromRoute();
  const { data } = useQuery({ ...cyclesListQuery(team.id), enabled: team.cycles_enabled });
  const latest = [...(data?.cycles ?? [])].sort((a, b) => b.ends_at.localeCompare(a.ends_at))[0];

  return (
    <PageContainer>
      <PageHeader
        title="Cycles"
        icon={<RefreshCwIcon />}
        breadcrumb={[teamCrumb(team, workspace.slug), { label: "Cycles" }]}
        actions={
          isAdmin &&
          team.cycles_enabled && (
            <CycleDialog
              mode="create"
              team={team}
              after={latest}
              trigger={
                <Button>
                  <PlusIcon className="size-4" />
                  New cycle
                </Button>
              }
            />
          )
        }
      />
      {team.cycles_enabled ? (
        <CycleList team={team} workspaceSlug={workspace.slug} />
      ) : (
        <EmptyState
          icon={RefreshCwIcon}
          title="Cycles are off for this team"
          description="Cycles are fixed time boxes; unfinished issues roll over to the next one."
          action={
            isAdmin && (
              <LinkButton
                to="/$workspace/settings/team/$key/cycles"
                params={{ workspace: workspace.slug, key: team.key }}
                variant="outline"
                size="sm"
              >
                Turn on cycles
              </LinkButton>
            )
          }
        />
      )}
    </PageContainer>
  );
}
