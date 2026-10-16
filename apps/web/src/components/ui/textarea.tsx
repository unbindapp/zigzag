import * as React from "react";

import { cn } from "@/components/ui/utils";

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "bg-input placeholder:text-muted-foreground/9-10 has-hover:hover:ring-primary/6-10 focus-visible:ring-primary/8-10 aria-invalid:border-destructive/8-10 flex min-h-20 w-full rounded-md border px-3 py-2 text-sm leading-normal focus-visible:ring-1 focus-visible:outline-hidden disabled:cursor-not-allowed disabled:opacity-50 has-hover:hover:ring-1",
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };
