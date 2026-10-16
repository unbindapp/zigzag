import { ReactNode } from "react";

import { cn } from "@/components/ui/utils";

export function List({ children }: { children: ReactNode }) {
  return (
    <div
      data-slot="list"
      role="list"
      className="bg-card flex w-full flex-col overflow-hidden rounded-lg border"
    >
      {children}
    </div>
  );
}

type TListRowProps = {
  children: ReactNode;
  actions?: ReactNode;
  muted?: boolean;
};

export function ListRow({ children, actions, muted }: TListRowProps) {
  return (
    <div
      role="listitem"
      data-slot="list-row"
      className={cn(
        "flex h-9 w-full items-center gap-2 border-b px-3 text-sm last:border-b-0",
        muted && "text-muted-foreground",
      )}
    >
      <div className="flex min-w-0 flex-1 items-center gap-2">{children}</div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </div>
  );
}

export function ListRowText({ primary, secondary }: { primary: ReactNode; secondary?: ReactNode }) {
  return (
    <div className="flex min-w-0 flex-1 items-baseline gap-2">
      <span className="truncate">{primary}</span>
      {secondary && <span className="text-muted-foreground truncate text-xs">{secondary}</span>}
    </div>
  );
}
