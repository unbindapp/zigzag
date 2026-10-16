import { RotateCwIcon, TriangleAlertIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/components/ui/utils";

type TProps = {
  title?: string;
  message?: string;
  onRetry?: () => void;
  className?: string;
};

export function ErrorCard({ title = "Something went wrong", message, onRetry, className }: TProps) {
  return (
    <div
      role="alert"
      data-slot="error-card"
      className={cn(
        "border-destructive/5-10 bg-destructive/1-10 flex w-full items-start gap-2 rounded-lg border p-4",
        className,
      )}
    >
      <TriangleAlertIcon className="text-destructive mt-0.5 size-4 shrink-0" />
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <p className="text-destructive text-sm font-medium">{title}</p>
        {message && <p className="text-muted-foreground font-mono text-xs break-all">{message}</p>}
      </div>
      {onRetry && (
        <Button size="sm" variant="outline" onClick={onRetry} className="shrink-0">
          <RotateCwIcon className="size-3.5" />
          Retry
        </Button>
      )}
    </div>
  );
}
