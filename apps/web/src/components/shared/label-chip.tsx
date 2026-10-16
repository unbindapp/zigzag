type TProps = {
  name: string;
  color?: string | null;
};

export function LabelChip({ name, color }: TProps) {
  return (
    <span className="bg-background text-foreground inline-flex h-5 max-w-40 shrink-0 items-center gap-1.5 rounded-full border px-2 text-xs leading-none">
      <span
        aria-hidden
        className="bg-muted-foreground size-2 shrink-0 rounded-full"
        style={color ? { backgroundColor: color } : undefined}
      />
      <span className="truncate">{name}</span>
    </span>
  );
}
