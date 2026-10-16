import * as React from "react";

import { cn } from "@/components/ui/utils";

function Skeleton({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="skeleton"
      aria-hidden
      className={cn("bg-foreground animate-skeleton rounded-md", className)}
      {...props}
    />
  );
}

export { Skeleton };
