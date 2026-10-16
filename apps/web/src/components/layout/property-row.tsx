import { ReactNode } from "react";

type TProps = {
  label: string;
  children: ReactNode;
};

export function PropertyRow({ label, children }: TProps) {
  return (
    <div data-slot="property-row" className="flex min-h-8 w-full items-center gap-2">
      <span className="text-muted-foreground w-24 shrink-0 truncate text-sm">{label}</span>
      <div className="flex min-w-0 flex-1 items-center gap-1.5 text-sm">{children}</div>
    </div>
  );
}
