import * as React from "react";

import { cn } from "@/components/ui/utils";

function Kbd({ className, ...props }: React.ComponentProps<"kbd">) {
  return (
    <kbd
      data-slot="kbd"
      className={cn(
        "bg-muted text-muted-foreground inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded-sm border px-1 font-sans text-xs leading-none font-medium",
        className,
      )}
      {...props}
    />
  );
}

function KbdGroup({ className, ...props }: React.ComponentProps<"span">) {
  return <span className={cn("inline-flex items-center gap-0.5", className)} {...props} />;
}

export { Kbd, KbdGroup };
