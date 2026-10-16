import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { teamKeyPattern } from "@/lib/helpers/identifiers";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TTeam = TResponseData<TGoClient["teams"]["get"]>;
export type TTeamMember = TListItem<TGoClient["teams"]["members"]["list"]>;
export type TWorkflowState = TListItem<TGoClient["teams"]["states"]["list"]>;
export type TCreateTeamInput = TInput<TGoClient["teams"]["create"]>;
export type TUpdateTeamInput = TInput<TGoClient["teams"]["update"]>;
export type TCreateWorkflowStateInput = TInput<TGoClient["teams"]["states"]["create"]>;
export type TUpdateWorkflowStateInput = TInput<TGoClient["teams"]["states"]["update"]>;

export const queryKeyTeams = {
  all: () => ["teams"] as const,
  list: (workspaceId: string) => ["teams", "list", workspaceId] as const,
  detail: (workspaceId: string, key: string) => ["teams", "detail", workspaceId, key] as const,
  members: (teamId: string) => ["teams", "members", teamId] as const,
  states: (teamId: string) => ["teams", "states", teamId] as const,
};

export const teamsListQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyTeams.list(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().teams.list({ workspace_id: workspaceId });
      return { teams: res.data ?? [] };
    },
  });

export const teamByKeyQuery = (workspaceId: string, key: string) =>
  queryOptions({
    queryKey: queryKeyTeams.detail(workspaceId, key),
    queryFn: async () => {
      const res = await getGoClient().teams.get({ workspace_id: workspaceId, key });
      return { team: res.data };
    },
  });

export const teamMembersQuery = (teamId: string) =>
  queryOptions({
    queryKey: queryKeyTeams.members(teamId),
    queryFn: async () => {
      const res = await getGoClient().teams.members.list({ team_id: teamId });
      return { members: res.data ?? [] };
    },
  });

export const workflowStatesQuery = (teamId: string) =>
  queryOptions({
    queryKey: queryKeyTeams.states(teamId),
    queryFn: async () => {
      const res = await getGoClient().teams.states.list({ team_id: teamId });
      return { states: res.data ?? [] };
    },
  });

export async function createTeam(input: TCreateTeamInput) {
  const res = await getGoClient().teams.create(input);
  return { data: res.data };
}

export async function updateTeam(input: TUpdateTeamInput) {
  const res = await getGoClient().teams.update(input);
  return { data: res.data };
}

export async function deleteTeam(input: { teamId: string }) {
  const res = await getGoClient().teams.delete({ team_id: input.teamId });
  return { data: res.data };
}

export async function addTeamMember(input: { teamId: string; userId: string; owner?: boolean }) {
  const res = await getGoClient().teams.members.add({
    team_id: input.teamId,
    user_id: input.userId,
    owner: input.owner,
  });
  return { data: res.data };
}

export async function removeTeamMember(input: { teamId: string; userId: string }) {
  const res = await getGoClient().teams.members.remove({
    team_id: input.teamId,
    user_id: input.userId,
  });
  return { data: res.data };
}

export async function createWorkflowState(input: TCreateWorkflowStateInput) {
  const res = await getGoClient().teams.states.create(input);
  return { data: res.data };
}

export async function updateWorkflowState(input: TUpdateWorkflowStateInput) {
  const res = await getGoClient().teams.states.update(input);
  return { data: res.data };
}

export async function deleteWorkflowState(input: { stateId: string; moveIssuesToStateId: string }) {
  const res = await getGoClient().teams.states.delete({
    state_id: input.stateId,
    move_issues_to_state_id: input.moveIssuesToStateId,
  });
  return { data: res.data };
}

function useInvalidateTeams() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyTeams.all() });
}

export function useCreateTeamMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: createTeam,
    meta: { errorTitle: "Could not create team" },
    onSuccess: invalidate,
  });
}

export function useUpdateTeamMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: updateTeam,
    meta: { errorTitle: "Could not update team" },
    onSuccess: invalidate,
  });
}

export function useDeleteTeamMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteTeam,
    meta: { errorTitle: "Could not delete team" },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeyTeams.all() });
      await queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}

export function useAddTeamMemberMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: addTeamMember,
    meta: { errorTitle: "Could not add member" },
    onSuccess: invalidate,
  });
}

export function useRemoveTeamMemberMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: removeTeamMember,
    meta: { errorTitle: "Could not remove member" },
    onSuccess: invalidate,
  });
}

export function useCreateWorkflowStateMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: createWorkflowState,
    meta: { errorTitle: "Could not add state" },
    onSuccess: invalidate,
  });
}

export function useUpdateWorkflowStateMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: updateWorkflowState,
    meta: { errorTitle: "Could not update state" },
    onSuccess: invalidate,
  });
}

export function useDeleteWorkflowStateMutation() {
  const invalidate = useInvalidateTeams();
  return useMutation({
    mutationFn: deleteWorkflowState,
    meta: { errorTitle: "Could not delete state" },
    onSuccess: invalidate,
  });
}

export const teamNameMaxLength = 64;

export const TeamFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(teamNameMaxLength, `Name should be at most ${teamNameMaxLength} characters.`),
  key: z
    .string()
    .trim()
    .regex(teamKeyPattern, "Start with a letter, then up to 6 letters or digits."),
});
