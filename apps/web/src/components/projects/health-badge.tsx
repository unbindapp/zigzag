import { CircleAlertIcon, CircleCheckIcon, CircleXIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import type { THealth } from "@/lib/queries/projects";

export const HEALTHS: THealth[] = ["on_track", "at_risk", "off_track"];

export const healthLabels: Record<THealth, string> = {
  on_track: "On track",
  at_risk: "At risk",
  off_track: "Off track",
};

const healthVariants = {
  on_track: "success",
  at_risk: "warning",
  off_track: "destructive",
} as const;

export function HealthIcon({ health }: { health: THealth }) {
  switch (health) {
    case "on_track":
      return <CircleCheckIcon className="text-success size-4" aria-hidden />;
    case "at_risk":
      return <CircleAlertIcon className="text-warning size-4" aria-hidden />;
    case "off_track":
      return <CircleXIcon className="text-destructive size-4" aria-hidden />;
  }
}

export function HealthBadge({ health }: { health: THealth | null | undefined }) {
  if (!health) return <Badge variant="muted">No updates</Badge>;
  return (
    <Badge variant={healthVariants[health]}>
      <HealthIcon health={health} />
      {healthLabels[health]}
    </Badge>
  );
}
