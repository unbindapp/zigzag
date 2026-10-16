import { formatDate } from "@/lib/helpers/dates";
import type { TCycle } from "@/lib/queries/cycles";

export type TCycleStatus = "active" | "upcoming" | "past";

export function cycleStatus(cycle: TCycle): TCycleStatus {
  if (cycle.is_active) return "active";
  if (cycle.is_future) return "upcoming";
  return "past";
}

export const cycleStatusLabels: Record<TCycleStatus, string> = {
  active: "Active",
  upcoming: "Upcoming",
  past: "Past",
};

export function cycleDateRange(cycle: { starts_at: string; ends_at: string }) {
  const lastDay = new Date(new Date(cycle.ends_at).getTime() - 1);
  return `${formatDate(cycle.starts_at)} – ${formatDate(lastDay)}`;
}

export function formatScope(value: number) {
  return Number.isInteger(value) ? String(value) : value.toFixed(1);
}
