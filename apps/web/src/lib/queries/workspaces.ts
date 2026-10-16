import { type QueryClient, queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { workspaceSlugPattern } from "@/lib/helpers/identifiers";
import { queryKeyMe } from "@/lib/queries/me";
import { teamsListQuery } from "@/lib/queries/teams";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TWorkspace = TResponseData<TGoClient["workspaces"]["get"]>;
export type TWorkspaceRole = TWorkspace["role"];
export type TWorkspaceMember = TListItem<TGoClient["workspaces"]["members"]["list"]>;
export type TInvitation = TListItem<TGoClient["workspaces"]["invitations"]["list"]>;
export type TCreatedInvitation = TResponseData<TGoClient["workspaces"]["invitations"]["create"]>;
export type TCreateWorkspaceInput = TInput<TGoClient["workspaces"]["create"]>;
export type TUpdateWorkspaceInput = TInput<TGoClient["workspaces"]["update"]>;
export type TCreateInvitationInput = TInput<TGoClient["workspaces"]["invitations"]["create"]>;

export const NO_WORKSPACE_ICON_UPLOAD_ID = "00000000-0000-0000-0000-000000000000";

export const WORKSPACE_ROLES: TWorkspaceRole[] = ["admin", "member", "viewer"];

export const workspaceRoleLabels: Record<TWorkspaceRole, string> = {
  admin: "Admin",
  member: "Member",
  viewer: "Viewer",
};

export const queryKeyWorkspaces = {
  all: () => ["workspaces"] as const,
  list: () => ["workspaces", "list"] as const,
  detail: (slug: string) => ["workspaces", "detail", slug] as const,
  members: (workspaceId: string) => ["workspaces", "members", workspaceId] as const,
  invitations: (workspaceId: string) => ["workspaces", "invitations", workspaceId] as const,
};

export const workspacesListQuery = () =>
  queryOptions({
    queryKey: queryKeyWorkspaces.list(),
    queryFn: async () => {
      const res = await getGoClient().workspaces.list();
      return { workspaces: res.data ?? [] };
    },
  });

export const workspaceBySlugQuery = (slug: string) =>
  queryOptions({
    queryKey: queryKeyWorkspaces.detail(slug),
    queryFn: async () => {
      const res = await getGoClient().workspaces.get({ slug });
      return { workspace: res.data };
    },
  });

export const workspaceMembersQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyWorkspaces.members(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().workspaces.members.list({ workspace_id: workspaceId });
      return { members: res.data ?? [] };
    },
  });

export const invitationsListQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyWorkspaces.invitations(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().workspaces.invitations.list({ workspace_id: workspaceId });
      return { invitations: res.data ?? [] };
    },
  });

export async function fetchWorkspaceLandingPath(
  queryClient: QueryClient,
  workspace: Pick<TWorkspace, "id" | "slug" | "default_team_id">,
) {
  const { teams } = await queryClient.fetchQuery(teamsListQuery(workspace.id)).catch(() => ({
    teams: [],
  }));
  const team = teams.find((item) => item.id === workspace.default_team_id);
  if (!team) return `/${workspace.slug}`;
  return `/${workspace.slug}/team/${team.key}/issues`;
}

export async function createWorkspace(input: TCreateWorkspaceInput) {
  const res = await getGoClient().workspaces.create(input);
  return { data: res.data };
}

export async function updateWorkspace(input: TUpdateWorkspaceInput) {
  const res = await getGoClient().workspaces.update(input);
  return { data: res.data };
}

export async function deleteWorkspace(input: { workspaceId: string; confirmSlug: string }) {
  const res = await getGoClient().workspaces.delete({
    workspace_id: input.workspaceId,
    confirm_slug: input.confirmSlug,
  });
  return { data: res.data };
}

export async function updateWorkspaceMember(input: {
  workspaceId: string;
  userId: string;
  role: TWorkspaceRole;
}) {
  const res = await getGoClient().workspaces.members.update({
    workspace_id: input.workspaceId,
    user_id: input.userId,
    role: input.role,
  });
  return { data: res.data };
}

export async function removeWorkspaceMember(input: { workspaceId: string; userId: string }) {
  const res = await getGoClient().workspaces.members.remove({
    workspace_id: input.workspaceId,
    user_id: input.userId,
  });
  return { data: res.data };
}

export async function createInvitation(input: TCreateInvitationInput) {
  const res = await getGoClient().workspaces.invitations.create(input);
  return { data: res.data };
}

export async function revokeInvitation(input: { workspaceId: string; invitationId: string }) {
  const res = await getGoClient().workspaces.invitations.revoke({
    workspace_id: input.workspaceId,
    invitation_id: input.invitationId,
  });
  return { data: res.data };
}

export async function resendInvitation(input: { workspaceId: string; invitationId: string }) {
  const res = await getGoClient().workspaces.invitations.resend({
    workspace_id: input.workspaceId,
    invitation_id: input.invitationId,
  });
  return { data: res.data };
}

function useInvalidateWorkspaces() {
  const queryClient = useQueryClient();
  return async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeyWorkspaces.all() }),
      queryClient.invalidateQueries({ queryKey: queryKeyMe }),
    ]);
  };
}

export function useCreateWorkspaceMutation() {
  const invalidate = useInvalidateWorkspaces();
  return useMutation({
    mutationFn: createWorkspace,
    meta: { errorTitle: "Could not create workspace" },
    onSuccess: invalidate,
  });
}

export function useUpdateWorkspaceMutation() {
  const invalidate = useInvalidateWorkspaces();
  return useMutation({
    mutationFn: updateWorkspace,
    meta: { errorTitle: "Could not update workspace" },
    onSuccess: invalidate,
  });
}

export function useDeleteWorkspaceMutation() {
  const invalidate = useInvalidateWorkspaces();
  return useMutation({
    mutationFn: deleteWorkspace,
    meta: { errorTitle: "Could not delete workspace" },
    onSuccess: invalidate,
  });
}

export function useUpdateWorkspaceMemberMutation(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateWorkspaceMember,
    meta: { errorTitle: "Could not change role" },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeyWorkspaces.members(workspaceId) }),
  });
}

export function useRemoveWorkspaceMemberMutation(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: removeWorkspaceMember,
    meta: { errorTitle: "Could not remove member" },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeyWorkspaces.members(workspaceId) });
      await queryClient.invalidateQueries({ queryKey: ["teams"] });
    },
  });
}

export function useCreateInvitationMutation(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createInvitation,
    meta: { errorTitle: "Could not send invitation" },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeyWorkspaces.invitations(workspaceId) }),
  });
}

export function useRevokeInvitationMutation(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: revokeInvitation,
    meta: { errorTitle: "Could not revoke invitation" },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeyWorkspaces.invitations(workspaceId) }),
  });
}

export function useResendInvitationMutation(workspaceId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: resendInvitation,
    meta: { errorTitle: "Could not resend invitation" },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeyWorkspaces.invitations(workspaceId) }),
  });
}

export const workspaceNameMaxLength = 64;

export const WorkspaceFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(workspaceNameMaxLength, `Name should be at most ${workspaceNameMaxLength} characters.`),
  slug: z
    .string()
    .trim()
    .regex(workspaceSlugPattern, "Use 2 to 32 lowercase letters, numbers or dashes."),
});
