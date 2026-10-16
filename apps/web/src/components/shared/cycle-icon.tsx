import { cn } from "@/components/ui/utils";

type TProps = {
  progress?: number;
  status?: "active" | "upcoming" | "past";
  className?: string;
};

export function CycleIcon({ progress = 0, status = "upcoming", className }: TProps) {
  const circumference = 2 * Math.PI * 4.5;
  const filled = Math.min(1, Math.max(0, progress)) * circumference;
  return (
    <svg
      viewBox="0 0 14 14"
      role="img"
      aria-label={`Cycle, ${status}`}
      data-status={status}
      className={cn(
        "text-muted-foreground data-[status=active]:text-primary size-3.5 shrink-0",
        className,
      )}
    >
      <circle
        cx="7"
        cy="7"
        r="6"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeDasharray={status === "upcoming" ? "1.4 1.74" : undefined}
      />
      {filled > 0 && (
        <circle
          cx="7"
          cy="7"
          r="4.5"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          strokeDasharray={`${filled} ${circumference}`}
          transform="rotate(-90 7 7)"
          opacity="0.6"
        />
      )}
    </svg>
  );
}
