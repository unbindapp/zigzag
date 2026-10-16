import { useParams } from "@tanstack/react-router";

import { useTeams } from "@/components/workspace/teams-provider";

export function useTeamFromRoute() {
  const { key } = useParams({ strict: false });
  const { teamByKey } = useTeams();
  return key ? teamByKey(key) : undefined;
}

export function useRequiredTeamFromRoute() {
  const team = useTeamFromRoute();
  if (!team)
    throw new Error("useRequiredTeamFromRoute must be used below a route that resolves the team");
  return team;
}
