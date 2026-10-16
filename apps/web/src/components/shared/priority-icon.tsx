import { cn } from "@/components/ui/utils";

export const PRIORITIES = [0, 1, 2, 3, 4] as const;
export type TPriority = (typeof PRIORITIES)[number];

const labels: Record<number, string> = {
  0: "No priority",
  1: "Urgent",
  2: "High",
  3: "Medium",
  4: "Low",
};

export function priorityLabel(priority: number) {
  return labels[priority] ?? labels[0];
}

const colorClass: Record<number, string> = {
  0: "text-muted-foreground",
  1: "text-destructive",
  2: "text-warning",
  3: "text-process",
  4: "text-muted-foreground",
};

const filledBars: Record<number, number> = { 2: 3, 3: 2, 4: 1 };

type TProps = {
  priority: number;
  className?: string;
};

export function PriorityIcon({ priority, className }: TProps) {
  const label = priorityLabel(priority);
  const classes = cn("size-4 shrink-0", colorClass[priority] ?? colorClass[0], className);

  switch (priority) {
    case 1:
      return (
        <svg viewBox="0 0 16 16" role="img" aria-label={label} className={classes}>
          <rect x="1.5" y="1.5" width="13" height="13" rx="3" fill="currentColor" />
          <rect x="7.25" y="4" width="1.5" height="5" rx="0.75" className="fill-background" />
          <rect x="7.25" y="10.5" width="1.5" height="1.5" rx="0.75" className="fill-background" />
        </svg>
      );
    case 2:
    case 3:
    case 4:
      return (
        <svg viewBox="0 0 16 16" role="img" aria-label={label} className={classes}>
          {[0, 1, 2].map((index) => (
            <rect
              key={index}
              x={2 + index * 4.5}
              y={10 - index * 3.5}
              width="3"
              height={4 + index * 3.5}
              rx="1"
              fill="currentColor"
              opacity={index < filledBars[priority] ? 1 : 0.3}
            />
          ))}
        </svg>
      );
    default:
      return (
        <svg viewBox="0 0 16 16" role="img" aria-label={label} className={classes}>
          {[0, 1, 2].map((index) => (
            <rect
              key={index}
              x={2 + index * 4.5}
              y="7.25"
              width="3"
              height="1.5"
              rx="0.75"
              fill="currentColor"
            />
          ))}
        </svg>
      );
  }
}
