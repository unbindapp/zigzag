import { BurnUpChart } from "@/components/cycles/burn-up-chart";
import { CycleProgressBar } from "@/components/cycles/cycle-progress-bar";
import {
  cycleDateRange,
  cycleStatus,
  cycleStatusLabels,
  formatScope,
} from "@/components/cycles/cycle-status";
import { IssueList } from "@/components/issues/issue-list";
import { SectionTitle } from "@/components/layout/section-title";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { cycleName, type TCycleDetail } from "@/lib/queries/cycles";
import type { TTeam } from "@/lib/queries/teams";

type TProps = {
  cycle: TCycleDetail;
  team: TTeam;
  workspaceId: string;
};

export function CycleDetail({ cycle, team, workspaceId }: TProps) {
  const status = cycleStatus(cycle);
  const progress = cycle.progress;
  const unit = team.estimate_type === "none" ? "issues" : "points";
  return (
    <>
      <Card className="flex w-full flex-col gap-4 p-4">
        <div className="flex w-full flex-wrap items-center gap-2">
          <Badge variant={status === "active" ? "process" : "muted"}>
            {cycleStatusLabels[status]}
          </Badge>
          <span className="text-muted-foreground text-sm">{cycleDateRange(cycle)}</span>
          {cycle.archived_at && <Badge variant="muted">Archived</Badge>}
        </div>
        {cycle.description && <p className="text-sm">{cycle.description}</p>}
        <div className="flex w-full items-center gap-2">
          <CycleProgressBar percent={progress.percent} label={`${cycleName(cycle)} progress`} />
          <span className="text-muted-foreground w-10 shrink-0 text-right text-xs">
            {Math.round(progress.percent)}%
          </span>
        </div>
        <dl className="grid w-full grid-cols-2 gap-3 sm:grid-cols-4">
          <ScopeNumber label="Scope" value={`${formatScope(progress.scope)} ${unit}`} />
          <ScopeNumber label="Started" value={formatScope(progress.started)} />
          <ScopeNumber label="Completed" value={formatScope(progress.completed)} />
          <ScopeNumber
            label="Issues done"
            value={`${progress.completed_issue_count} of ${progress.issue_count}`}
          />
        </dl>
        <BurnUpChart cycle={cycle} />
      </Card>
      <section className="flex w-full flex-col gap-2">
        <SectionTitle title="Issues" />
        <IssueList
          workspaceId={workspaceId}
          filters={{}}
          fixedFilters={{ cycle_ids: [cycle.id] }}
        />
      </section>
    </>
  );
}

function ScopeNumber({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <dt className="text-muted-foreground text-xs">{label}</dt>
      <dd className="truncate text-sm font-medium">{value}</dd>
    </div>
  );
}
