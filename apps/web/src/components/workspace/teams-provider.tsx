import { useQuery } from "@tanstack/react-query";
import { createContext, ReactNode, useContext, useMemo } from "react";

import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TTeam, teamsListQuery } from "@/lib/queries/teams";

type TTeamsContext = {
  teams: TTeam[];
  myTeams: TTeam[];
  isPending: boolean;
  error: Error | null;
  refetch: () => void;
  teamByKey: (key: string) => TTeam | undefined;
};

const TeamsContext = createContext<TTeamsContext | null>(null);

export function TeamsProvider({ children }: { children: ReactNode }) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(teamsListQuery(workspace.id));
  const value = useMemo<TTeamsContext>(() => {
    const teams = data?.teams ?? [];
    return {
      teams,
      myTeams: teams.filter((team) => team.is_member),
      isPending,
      error,
      refetch: () => void refetch(),
      teamByKey: (key: string) =>
        teams.find((team) => team.key.toLowerCase() === key.toLowerCase()),
    };
  }, [data, isPending, error, refetch]);
  return <TeamsContext.Provider value={value}>{children}</TeamsContext.Provider>;
}

export function useTeams() {
  const context = useContext(TeamsContext);
  if (!context) throw new Error("useTeams must be used within a TeamsProvider");
  return context;
}
