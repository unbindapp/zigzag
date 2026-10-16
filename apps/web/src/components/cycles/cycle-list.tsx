import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { RefreshCwIcon } from "lucide-react";

import { CycleProgressBar } from "@/components/cycles/cycle-progress-bar";
import {
  cycleDateRange,
  cycleStatus,
  cycleStatusLabels,
  formatScope,
  type TCycleStatus,
} from "@/components/cycles/cycle-status";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SectionTitle } from "@/components/layout/section-title";
import { CycleIcon } from "@/components/shared/cycle-icon";
import { Badge } from "@/components/ui/badge";
import { cycleName, cyclesListQuery, type TCycle } from "@/lib/queries/cycles";
import type { TTeam } from "@/lib/queries/teams";

const SECTIONS: TCycleStatus[] = ["active", "upcoming", "past"];

type TProps = {
  team: TTeam;
  workspaceSlug: string;
};

export function CycleList({ team, workspaceSlug }: TProps) {
  const { data, isPending, error, refetch } = useQuery(cyclesListQuery(team.id));
  const cycles = data?.cycles ?? [];
  return (
    <ListStates
      isPending={isPending}
      error={error}
      onRetry={() => void refetch()}
      isEmpty={cycles.length === 0}
      empty={{
        icon: RefreshCwIcon,
        title: "No cycles yet",
        description:
          "Cycles are planned automatically every few minutes once cycles are on. Admins can also create one now.",
      }}
    >
      <div className="flex w-full flex-col gap-6">
        {SECTIONS.map((status) => {
          const inSection = cycles.filter((cycle) => cycleStatus(cycle) === status);
          if (inSection.length === 0) return null;
          const ordered = status === "upcoming" ? [...inSection].reverse() : inSection;
          return (
            <section key={status} className="flex w-full flex-col gap-2">
              <SectionTitle title={cycleStatusLabels[status]} />
              <List>
                {ordered.map((cycle) => (
                  <CycleRow
                    key={cycle.id}
                    cycle={cycle}
                    team={team}
                    workspaceSlug={workspaceSlug}
                  />
                ))}
              </List>
            </section>
          );
        })}
      </div>
    </ListStates>
  );
}

function CycleRow({ cycle, team, workspaceSlug }: { cycle: TCycle } & TProps) {
  const status = cycleStatus(cycle);
  const progress = cycle.progress;
  const unit = team.estimate_type === "none" ? "issues" : "points";
  return (
    <Link
      to="/$workspace/team/$key/cycle/$number"
      params={{ workspace: workspaceSlug, key: team.key, number: String(cycle.number) }}
      role="listitem"
      className="hover:bg-accent focus-visible:bg-accent flex h-9 w-full items-center gap-2 border-b px-3 text-sm outline-hidden last:border-b-0"
    >
      <CycleIcon progress={progress.percent / 100} status={status} className="size-4" />
      <span className="min-w-0 flex-1 truncate">{cycleName(cycle)}</span>
      {status === "active" && <Badge variant="process">Current</Badge>}
      <span className="text-muted-foreground hidden text-xs sm:inline">
        {cycleDateRange(cycle)}
      </span>
      <span className="hidden w-24 md:flex">
        <CycleProgressBar percent={progress.percent} label={`${cycleName(cycle)} progress`} />
      </span>
      <span className="text-muted-foreground w-28 text-right text-xs">
        {formatScope(progress.completed)} / {formatScope(progress.scope)} {unit}
      </span>
      <span className="text-muted-foreground w-10 text-right text-xs">
        {Math.round(progress.percent)}%
      </span>
    </Link>
  );
}
