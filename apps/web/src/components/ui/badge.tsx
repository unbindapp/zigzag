import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "@/components/ui/utils";

const badgeVariants = cva(
  "inline-flex h-5 max-w-full shrink-0 items-center gap-1 rounded-full border px-2 text-xs leading-none font-medium whitespace-nowrap [&_svg]:size-3.5 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground border-transparent",
        outline: "text-foreground bg-background",
        muted: "bg-muted text-muted-foreground border-transparent",
        destructive: "text-destructive bg-destructive/3-10 border-destructive/5-10",
        warning: "text-warning bg-warning/3-10 border-warning/5-10",
        success: "text-success bg-success/3-10 border-success/5-10",
        process: "text-process bg-process/3-10 border-process/5-10",
      },
    },
    defaultVariants: { variant: "outline" },
  },
);

function Badge({
  className,
  variant,
  ...props
}: React.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
  return (
    <span data-slot="badge" className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

export { Badge, badgeVariants };
