import { StateIcon, type TStateType } from "@/components/shared/state-icon";
import type { TProjectStatusType } from "@/lib/queries/projects";

export const PROJECT_STATUS_TYPES: TProjectStatusType[] = [
  "backlog",
  "planned",
  "started",
  "paused",
  "completed",
  "canceled",
];

export const projectStatusTypeLabels: Record<TProjectStatusType, string> = {
  backlog: "Backlog",
  planned: "Planned",
  started: "Started",
  paused: "Paused",
  completed: "Completed",
  canceled: "Canceled",
};

const iconTypes: Record<TProjectStatusType, TStateType> = {
  backlog: "backlog",
  planned: "unstarted",
  started: "started",
  paused: "started",
  completed: "completed",
  canceled: "canceled",
};

type TProps = {
  type: TProjectStatusType;
  color?: string | null;
  name?: string;
};

export function ProjectStatusIcon({ type, color, name }: TProps) {
  return (
    <StateIcon
      type={iconTypes[type]}
      color={color}
      name={name ?? projectStatusTypeLabels[type]}
      progress={type === "paused" ? 0.25 : 0.5}
    />
  );
}
