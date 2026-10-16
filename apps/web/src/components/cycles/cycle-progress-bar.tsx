type TProps = {
  percent: number;
  label: string;
};

export function CycleProgressBar({ percent, label }: TProps) {
  const clamped = Math.min(100, Math.max(0, percent));
  return (
    <div
      role="progressbar"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-valuenow={Math.round(clamped)}
      className="bg-muted h-1.5 w-full overflow-hidden rounded-full"
    >
      <div className="bg-primary h-full rounded-full" style={{ width: `${clamped}%` }} />
    </div>
  );
}
