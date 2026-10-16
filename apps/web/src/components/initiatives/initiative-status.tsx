import { StateIcon, type TStateType } from "@/components/shared/state-icon";
import { Badge } from "@/components/ui/badge";
import type { TInitiativeStatus } from "@/lib/queries/initiatives";

export const INITIATIVE_STATUSES: TInitiativeStatus[] = ["planned", "active", "completed"];

export const initiativeStatusLabels: Record<TInitiativeStatus, string> = {
  planned: "Planned",
  active: "Active",
  completed: "Completed",
};

const iconTypes: Record<TInitiativeStatus, TStateType> = {
  planned: "unstarted",
  active: "started",
  completed: "completed",
};

export function InitiativeStatusIcon({ status }: { status: TInitiativeStatus }) {
  return (
    <StateIcon
      type={iconTypes[status]}
      name={initiativeStatusLabels[status]}
      className={status === "planned" ? undefined : "text-primary"}
    />
  );
}

export function InitiativeStatusBadge({ status }: { status: TInitiativeStatus }) {
  return (
    <Badge variant="outline">
      <InitiativeStatusIcon status={status} />
      {initiativeStatusLabels[status]}
    </Badge>
  );
}
