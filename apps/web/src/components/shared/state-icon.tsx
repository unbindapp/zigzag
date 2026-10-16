import { cn } from "@/components/ui/utils";

export type TStateType = "triage" | "backlog" | "unstarted" | "started" | "completed" | "canceled";

export const STATE_TYPES: TStateType[] = [
  "triage",
  "backlog",
  "unstarted",
  "started",
  "completed",
  "canceled",
];

export const stateTypeLabels: Record<TStateType, string> = {
  triage: "Triage",
  backlog: "Backlog",
  unstarted: "Unstarted",
  started: "Started",
  completed: "Completed",
  canceled: "Canceled",
};

type TProps = {
  type: TStateType;
  color?: string | null;
  name?: string;
  progress?: number;
  className?: string;
};

function partialArc(progress: number) {
  const clamped = Math.min(0.99, Math.max(0.01, progress));
  const angle = clamped * 2 * Math.PI;
  const radius = 3.5;
  const x = 7 + radius * Math.sin(angle);
  const y = 7 - radius * Math.cos(angle);
  const largeArc = clamped > 0.5 ? 1 : 0;
  return `M7 7 L7 ${7 - radius} A${radius} ${radius} 0 ${largeArc} 1 ${x} ${y} Z`;
}

export function StateIcon({ type, color, name, progress = 0.5, className }: TProps) {
  const label = name ?? stateTypeLabels[type];
  const classes = cn("size-3.5 shrink-0", !color && "text-muted-foreground", className);
  const style = color ? { color } : undefined;

  switch (type) {
    case "triage":
      return (
        <svg viewBox="0 0 14 14" role="img" aria-label={label} className={classes} style={style}>
          <circle cx="7" cy="7" r="6" fill="none" stroke="currentColor" strokeWidth="1.5" />
          <path d="M7 3.5 L10.25 9.5 L3.75 9.5 Z" fill="currentColor" />
        </svg>
      );
    case "backlog":
      return (
        <svg viewBox="0 0 14 14" role="img" aria-label={label} className={classes} style={style}>
          <circle
            cx="7"
            cy="7"
            r="6"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeDasharray="1.4 1.74"
          />
        </svg>
      );
    case "unstarted":
      return (
        <svg viewBox="0 0 14 14" role="img" aria-label={label} className={classes} style={style}>
          <circle cx="7" cy="7" r="6" fill="none" stroke="currentColor" strokeWidth="1.5" />
        </svg>
      );
    case "started":
      return (
        <svg viewBox="0 0 14 14" role="img" aria-label={label} className={classes} style={style}>
          <circle cx="7" cy="7" r="6" fill="none" stroke="currentColor" strokeWidth="1.5" />
          <path d={partialArc(progress)} fill="currentColor" />
        </svg>
      );
    case "completed":
      return (
        <svg viewBox="0 0 14 14" role="img" aria-label={label} className={classes} style={style}>
          <circle cx="7" cy="7" r="6.5" fill="currentColor" />
          <path
            d="M4.25 7.25 L6.1 9 L9.75 5.25"
            fill="none"
            className="stroke-background"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      );
    case "canceled":
      return (
        <svg viewBox="0 0 14 14" role="img" aria-label={label} className={classes} style={style}>
          <circle cx="7" cy="7" r="6.5" fill="currentColor" />
          <path
            d="M4.75 4.75 L9.25 9.25 M9.25 4.75 L4.75 9.25"
            fill="none"
            className="stroke-background"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
        </svg>
      );
  }
}
