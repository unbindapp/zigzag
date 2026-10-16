import { addDays, addYears } from "date-fns";
import { z } from "zod";

export type TAccess = "all" | "specific";

export const accessItems: { value: TAccess; label: string }[] = [
  { value: "all", label: "All my workspaces" },
  { value: "specific", label: "Specific workspaces" },
];

export const expiryOptions = [
  { value: "7d", label: "7 days" },
  { value: "30d", label: "30 days" },
  { value: "90d", label: "90 days" },
  { value: "1y", label: "1 year" },
  { value: "never", label: "Never" },
] as const;

export type TExpiryValue = (typeof expiryOptions)[number]["value"];
export const defaultExpiry: TExpiryValue = "90d";

export function expiresAtFrom(value: TExpiryValue, now = new Date()): string | undefined {
  switch (value) {
    case "7d":
      return addDays(now, 7).toISOString();
    case "30d":
      return addDays(now, 30).toISOString();
    case "90d":
      return addDays(now, 90).toISOString();
    case "1y":
      return addYears(now, 1).toISOString();
    case "never":
      return undefined;
  }
}

export const accessFormShape = {
  role: z.enum(["admin", "member", "viewer"]),
  access: z.enum(["all", "specific"]),
  workspace_ids: z.array(z.string()),
};

export function hasPickedWorkspace(value: { access: TAccess; workspace_ids: string[] }) {
  return value.access === "all" || value.workspace_ids.length > 0;
}

export const pickWorkspaceMessage = {
  message: "Pick at least one workspace.",
  path: ["workspace_ids"],
};

export function describeAccess(
  access: { full_access: boolean; workspace_ids?: string[] | null },
  workspaceNames: Record<string, string>,
) {
  if (access.full_access) return "All workspaces";
  const names = (access.workspace_ids ?? []).map((id) => workspaceNames[id] ?? "Unknown workspace");
  if (names.length === 0) return "No workspaces";
  return names.join(", ");
}
