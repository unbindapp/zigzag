type TProps = {
  percent: number;
  label: string;
  showValue?: boolean;
};

export function ProgressBar({ percent, label, showValue = true }: TProps) {
  const clamped = Math.min(100, Math.max(0, Math.round(percent)));
  return (
    <div className="flex min-w-0 items-center gap-2">
      <div
        role="progressbar"
        aria-label={label}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={clamped}
        className="bg-muted h-1.5 w-full min-w-12 overflow-hidden rounded-full"
      >
        <div className="bg-primary h-full rounded-full" style={{ width: `${clamped}%` }} />
      </div>
      {showValue && (
        <span className="text-muted-foreground w-9 shrink-0 text-right text-xs">{clamped}%</span>
      )}
    </div>
  );
}
