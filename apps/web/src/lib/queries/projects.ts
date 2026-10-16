import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { LIST_REFETCH_INTERVAL_MS } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TProject = TResponseData<TGoClient["projects"]["get"]>;
export type TProjectListItem = TListItem<TGoClient["projects"]["list"]>;
export type TProjectStatus = TListItem<TGoClient["projects"]["statuses"]["list"]>;
export type TProjectStatusType = TProjectStatus["type"];
export type TMilestone = TListItem<TGoClient["projects"]["milestones"]["list"]>;
export type TProjectUpdate = TListItem<TGoClient["projects"]["updates"]["list"]>;
export type THealth = TProjectUpdate["health"];
export type TProjectLink = TProject["links"][number];
export type TDateResolution = TProject["start_date_resolution"];
export type TListProjectsInput = TInput<TGoClient["projects"]["list"]>;
export type TCreateProjectInput = TInput<TGoClient["projects"]["create"]>;
export type TUpdateProjectInput = TInput<TGoClient["projects"]["update"]>;
export type TCreateProjectStatusInput = TInput<TGoClient["projects"]["statuses"]["create"]>;
export type TUpdateProjectStatusInput = TInput<TGoClient["projects"]["statuses"]["update"]>;
export type TCreateMilestoneInput = TInput<TGoClient["projects"]["milestones"]["create"]>;
export type TUpdateMilestoneInput = TInput<TGoClient["projects"]["milestones"]["update"]>;
export type TCreateProjectUpdateInput = TInput<TGoClient["projects"]["updates"]["create"]>;
export type TEditProjectUpdateInput = TInput<TGoClient["projects"]["updates"]["update"]>;
export type TCreateProjectLinkInput = TInput<TGoClient["projects"]["links"]["create"]>;
export type TUpdateProjectLinkInput = TInput<TGoClient["projects"]["links"]["update"]>;

export const queryKeyProjects = {
  all: () => ["projects"] as const,
  lists: () => ["projects", "list"] as const,
  list: (input: TListProjectsInput) => ["projects", "list", input] as const,
  detail: (workspaceId: string, slug: string) => ["projects", "detail", workspaceId, slug] as const,
  statuses: (workspaceId: string) => ["projects", "statuses", workspaceId] as const,
  updates: (projectId: string) => ["projects", "updates", projectId] as const,
};

export const projectsListQuery = (input: TListProjectsInput) =>
  queryOptions({
    queryKey: queryKeyProjects.list(input),
    queryFn: async () => {
      const res = await getGoClient().projects.list(input);
      return { projects: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const projectBySlugQuery = (workspaceId: string, slug: string) =>
  queryOptions({
    queryKey: queryKeyProjects.detail(workspaceId, slug),
    queryFn: async () => {
      const res = await getGoClient().projects.get({ workspace_id: workspaceId, slug });
      return { project: res.data };
    },
  });

export const projectStatusesQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyProjects.statuses(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().projects.statuses.list({ workspace_id: workspaceId });
      return { statuses: res.data ?? [] };
    },
  });

export const projectUpdatesQuery = (projectId: string) =>
  queryOptions({
    queryKey: queryKeyProjects.updates(projectId),
    queryFn: async () => {
      const res = await getGoClient().projects.updates.list({ project_id: projectId });
      return { updates: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export async function createProject(input: TCreateProjectInput) {
  const res = await getGoClient().projects.create(input);
  return { data: res.data };
}

export async function updateProject(input: TUpdateProjectInput) {
  const res = await getGoClient().projects.update(input);
  return { data: res.data };
}

export type TProjectLifecycleAction = "archive" | "unarchive" | "trash" | "restore" | "delete";

export async function changeProjectLifecycle(input: {
  projectId: string;
  action: TProjectLifecycleAction;
}) {
  const projects = getGoClient().projects;
  const params = { project_id: input.projectId };
  switch (input.action) {
    case "archive":
      return { data: (await projects.archive(params)).data };
    case "unarchive":
      return { data: (await projects.unarchive(params)).data };
    case "trash":
      return { data: (await projects.trash(params)).data };
    case "restore":
      return { data: (await projects.restore(params)).data };
    case "delete":
      await projects.delete(params);
      return { data: null };
  }
}

export async function createProjectStatus(input: TCreateProjectStatusInput) {
  const res = await getGoClient().projects.statuses.create(input);
  return { data: res.data };
}

export async function updateProjectStatus(input: TUpdateProjectStatusInput) {
  const res = await getGoClient().projects.statuses.update(input);
  return { data: res.data };
}

export async function deleteProjectStatus(input: { statusId: string; moveToStatusId: string }) {
  const res = await getGoClient().projects.statuses.delete({
    status_id: input.statusId,
    move_to_status_id: input.moveToStatusId,
  });
  return { data: res.data };
}

export async function createMilestone(input: TCreateMilestoneInput) {
  const res = await getGoClient().projects.milestones.create(input);
  return { data: res.data };
}

export async function updateMilestone(input: TUpdateMilestoneInput) {
  const res = await getGoClient().projects.milestones.update(input);
  return { data: res.data };
}

export async function deleteMilestone(input: { milestoneId: string }) {
  const res = await getGoClient().projects.milestones.delete({ milestone_id: input.milestoneId });
  return { data: res.data };
}

export async function createProjectUpdate(input: TCreateProjectUpdateInput) {
  const res = await getGoClient().projects.updates.create(input);
  return { data: res.data };
}

export async function editProjectUpdate(input: TEditProjectUpdateInput) {
  const res = await getGoClient().projects.updates.update(input);
  return { data: res.data };
}

export async function deleteProjectUpdate(input: { updateId: string }) {
  const res = await getGoClient().projects.updates.delete({ update_id: input.updateId });
  return { data: res.data };
}

export async function createProjectLink(input: TCreateProjectLinkInput) {
  const res = await getGoClient().projects.links.create(input);
  return { data: res.data };
}

export async function updateProjectLink(input: TUpdateProjectLinkInput) {
  const res = await getGoClient().projects.links.update(input);
  return { data: res.data };
}

export async function deleteProjectLink(input: { linkId: string }) {
  const res = await getGoClient().projects.links.delete({ link_id: input.linkId });
  return { data: res.data };
}

export async function toggleProjectUpdateReaction(input: {
  updateId: string;
  emoji: string;
  reacted: boolean;
}) {
  const reactions = getGoClient().issues.reactions;
  const call = input.reacted ? reactions.remove : reactions.add;
  const res = await call({ emoji: input.emoji, project_update_id: input.updateId });
  return { data: res.data };
}

function useInvalidateProjects() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyProjects.all() });
    await queryClient.invalidateQueries({ queryKey: ["initiatives"] });
  };
}

export function useCreateProjectMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: createProject,
    meta: { errorTitle: "Could not create project" },
    onSuccess: invalidate,
  });
}

export function useUpdateProjectMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: updateProject,
    meta: { errorTitle: "Could not update project" },
    onSuccess: invalidate,
  });
}

export function useProjectLifecycleMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: changeProjectLifecycle,
    meta: { errorTitle: "Could not change the project" },
    onSuccess: invalidate,
  });
}

export function useCreateProjectStatusMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: createProjectStatus,
    meta: { errorTitle: "Could not add status" },
    onSuccess: invalidate,
  });
}

export function useUpdateProjectStatusMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: updateProjectStatus,
    meta: { errorTitle: "Could not update status" },
    onSuccess: invalidate,
  });
}

export function useDeleteProjectStatusMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: deleteProjectStatus,
    meta: { errorTitle: "Could not delete status" },
    onSuccess: invalidate,
  });
}

export function useCreateMilestoneMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: createMilestone,
    meta: { errorTitle: "Could not add milestone" },
    onSuccess: invalidate,
  });
}

export function useUpdateMilestoneMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: updateMilestone,
    meta: { errorTitle: "Could not update milestone" },
    onSuccess: invalidate,
  });
}

export function useDeleteMilestoneMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteMilestone,
    meta: { errorTitle: "Could not delete milestone" },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeyProjects.all() });
      await queryClient.invalidateQueries({ queryKey: ["issues"] });
    },
  });
}

export function useCreateProjectUpdateMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: createProjectUpdate,
    meta: { errorTitle: "Could not post update" },
    onSuccess: invalidate,
  });
}

export function useEditProjectUpdateMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: editProjectUpdate,
    meta: { errorTitle: "Could not save update" },
    onSuccess: invalidate,
  });
}

export function useDeleteProjectUpdateMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: deleteProjectUpdate,
    meta: { errorTitle: "Could not delete update" },
    onSuccess: invalidate,
  });
}

export function useProjectUpdateReactionMutation(projectId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: toggleProjectUpdateReaction,
    meta: { errorTitle: "Could not react" },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeyProjects.updates(projectId) }),
  });
}

export function useCreateProjectLinkMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: createProjectLink,
    meta: { errorTitle: "Could not add link" },
    onSuccess: invalidate,
  });
}

export function useUpdateProjectLinkMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: updateProjectLink,
    meta: { errorTitle: "Could not update link" },
    onSuccess: invalidate,
  });
}

export function useDeleteProjectLinkMutation() {
  const invalidate = useInvalidateProjects();
  return useMutation({
    mutationFn: deleteProjectLink,
    meta: { errorTitle: "Could not delete link" },
    onSuccess: invalidate,
  });
}

export const projectNameMaxLength = 80;

export const ProjectFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(projectNameMaxLength, `Name should be at most ${projectNameMaxLength} characters.`),
  summary: z.string().trim().max(255, "Keep the summary under 255 characters."),
});
