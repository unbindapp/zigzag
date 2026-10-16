import { useQuery } from "@tanstack/react-query";

import { sortStates } from "@/components/issues/grouping";
import type { TUserSummary } from "@/lib/queries/issues";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { cyclesListQuery } from "@/lib/queries/cycles";
import { labelTreeQuery } from "@/lib/queries/labels";
import { projectsListQuery } from "@/lib/queries/projects";
import { workspaceMembersQuery } from "@/lib/queries/workspaces";

export function useTeam(teamId: string | undefined) {
  const { teams } = useTeams();
  return teams.find((team) => team.id === teamId);
}

export function useTeamStates(teamId: string | undefined) {
  const team = useTeam(teamId);
  return team ? sortStates(team.states) : [];
}

export function useWorkspaceUsers() {
  const { workspace } = useWorkspace();
  const { data, isPending } = useQuery(workspaceMembersQuery(workspace.id));
  const users: TUserSummary[] = (data?.members ?? [])
    .filter((member) => member.user.active)
    .map((member) => ({
      id: member.user.id,
      name: member.user.name,
      display_name: member.user.display_name,
      avatar_url: member.user.avatar_url,
    }));
  return { users, isPending };
}

export function useLabels(teamId: string | undefined) {
  const { workspace } = useWorkspace();
  const { data, isPending } = useQuery(
    labelTreeQuery({ workspaceId: workspace.id, teamId, allTeams: !teamId }),
  );
  return { labels: data?.labels ?? [], isPending };
}

function useWorkspaceProjects() {
  const { workspace } = useWorkspace();
  return useQuery(projectsListQuery({ workspace_id: workspace.id }));
}

export function useProjects(teamId: string | undefined) {
  const { data, isPending } = useWorkspaceProjects();
  const projects = (data?.projects ?? []).filter(
    (project) =>
      !teamId || project.teams.length === 0 || project.teams.some((team) => team.id === teamId),
  );
  return { projects, isPending };
}

export function useMilestones(projectId: string | undefined) {
  const { data, isPending } = useWorkspaceProjects();
  const project = data?.projects.find((item) => item.id === projectId);
  return { milestones: project?.milestones ?? [], isPending: Boolean(projectId) && isPending };
}

export function useCycles(teamId: string | undefined) {
  const team = useTeam(teamId);
  const enabled = Boolean(teamId && team?.cycles_enabled);
  const { data, isPending } = useQuery({ ...cyclesListQuery(teamId ?? ""), enabled });
  const cycles = (data?.cycles ?? []).filter((cycle) => !cycle.is_past);
  return { cycles, isPending: enabled && isPending, enabled };
}
