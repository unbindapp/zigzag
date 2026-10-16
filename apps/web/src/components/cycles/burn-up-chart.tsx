import { format } from "date-fns";

import { formatScope } from "@/components/cycles/cycle-status";
import type { TCycleDetail, TCycleScopePoint } from "@/lib/queries/cycles";

const DAY_MS = 24 * 60 * 60 * 1000;
const CHART_HEIGHT = 100;
const BAR_GAP = 0.2;

type TDay = { key: string; point?: TCycleScopePoint };

function dayKey(date: Date) {
  return format(date, "yyyy-MM-dd");
}

function cycleDays(cycle: TCycleDetail): TDay[] {
  const byDate = new Map((cycle.scope_history ?? []).map((point) => [point.date, point]));
  const start = new Date(cycle.starts_at);
  const count = Math.max(
    1,
    Math.round((new Date(cycle.ends_at).getTime() - start.getTime()) / DAY_MS),
  );
  return Array.from({ length: count }, (_, index) => {
    const key = dayKey(new Date(start.getTime() + index * DAY_MS));
    return { key, point: byDate.get(key) };
  });
}

export function BurnUpChart({ cycle }: { cycle: TCycleDetail }) {
  const days = cycleDays(cycle);
  const maxScope = Math.max(1, ...days.map((day) => day.point?.scope ?? 0));
  const barWidth = 1 - BAR_GAP;
  const scale = (value: number) => (value / maxScope) * CHART_HEIGHT;
  const last = [...days].reverse().find((day) => day.point)?.point;

  return (
    <figure className="flex w-full flex-col gap-2">
      <svg
        viewBox={`0 0 ${days.length} ${CHART_HEIGHT}`}
        preserveAspectRatio="none"
        role="img"
        aria-label={
          last
            ? `Scope ${formatScope(last.scope)}, started ${formatScope(last.started)}, completed ${formatScope(last.completed)}`
            : "No scope recorded yet"
        }
        className="h-32 w-full"
      >
        {days.map((day, index) => {
          if (!day.point) {
            return (
              <rect
                key={day.key}
                x={index + BAR_GAP / 2}
                y={CHART_HEIGHT - 1}
                width={barWidth}
                height={1}
                className="fill-border"
              />
            );
          }
          const scope = scale(day.point.scope);
          const done = scale(day.point.completed);
          const started = scale(day.point.started);
          return (
            <g key={day.key}>
              <title>{`${day.key}: scope ${formatScope(day.point.scope)}, started ${formatScope(day.point.started)}, completed ${formatScope(day.point.completed)}`}</title>
              <rect
                x={index + BAR_GAP / 2}
                y={CHART_HEIGHT - scope}
                width={barWidth}
                height={scope}
                className="fill-muted-foreground/25"
              />
              <rect
                x={index + BAR_GAP / 2}
                y={CHART_HEIGHT - done - started}
                width={barWidth}
                height={started}
                className="fill-warning"
              />
              <rect
                x={index + BAR_GAP / 2}
                y={CHART_HEIGHT - done}
                width={barWidth}
                height={done}
                className="fill-primary"
              />
            </g>
          );
        })}
      </svg>
      <figcaption className="text-muted-foreground flex items-center gap-2 text-xs">
        <span>{format(new Date(cycle.starts_at), "MMM d")}</span>
        <span className="flex flex-1 items-center justify-center gap-2">
          <LegendSwatch className="bg-muted-foreground/25" label="Scope" />
          <LegendSwatch className="bg-warning" label="Started" />
          <LegendSwatch className="bg-primary" label="Completed" />
        </span>
        <span>{format(new Date(new Date(cycle.ends_at).getTime() - 1), "MMM d")}</span>
      </figcaption>
    </figure>
  );
}

function LegendSwatch({ className, label }: { className: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5">
      <span aria-hidden className={`size-2 rounded-full ${className}`} />
      {label}
    </span>
  );
}
