import type { TBreadcrumbItem } from "@/components/layout/breadcrumb";
import { TeamIcon } from "@/components/shared/team-icon";
import type { TTeam } from "@/lib/queries/teams";

export function teamCrumb(
  team: Pick<TTeam, "key" | "name" | "icon" | "color">,
  workspaceSlug: string,
): TBreadcrumbItem {
  return {
    label: team.name,
    icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
    link: {
      to: "/$workspace/team/$key/issues",
      params: { workspace: workspaceSlug, key: team.key },
    },
  };
}
