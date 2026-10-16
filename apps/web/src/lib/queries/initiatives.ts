import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { LIST_REFETCH_INTERVAL_MS } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TInitiative = TResponseData<TGoClient["initiatives"]["get"]>;
export type TInitiativeListItem = TListItem<TGoClient["initiatives"]["list"]>;
export type TInitiativeStatus = TInitiative["status"];
export type TInitiativeUpdate = TListItem<TGoClient["initiatives"]["updates"]["list"]>;
export type TListInitiativesInput = TInput<TGoClient["initiatives"]["list"]>;
export type TCreateInitiativeInput = TInput<TGoClient["initiatives"]["create"]>;
export type TUpdateInitiativeInput = TInput<TGoClient["initiatives"]["update"]>;
export type TCreateInitiativeUpdateInput = TInput<TGoClient["initiatives"]["updates"]["create"]>;

export const NO_OWNER_ID = "00000000-0000-0000-0000-000000000000";

export const queryKeyInitiatives = {
  all: () => ["initiatives"] as const,
  lists: () => ["initiatives", "list"] as const,
  list: (input: TListInitiativesInput) => ["initiatives", "list", input] as const,
  detail: (workspaceId: string, slug: string) =>
    ["initiatives", "detail", workspaceId, slug] as const,
  updates: (initiativeId: string) => ["initiatives", "updates", initiativeId] as const,
};

export const initiativesListQuery = (input: TListInitiativesInput) =>
  queryOptions({
    queryKey: queryKeyInitiatives.list(input),
    queryFn: async () => {
      const res = await getGoClient().initiatives.list(input);
      return { initiatives: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const initiativeBySlugQuery = (workspaceId: string, slug: string) =>
  queryOptions({
    queryKey: queryKeyInitiatives.detail(workspaceId, slug),
    queryFn: async () => {
      const res = await getGoClient().initiatives.get({ workspace_id: workspaceId, slug });
      return { initiative: res.data };
    },
  });

export const initiativeUpdatesQuery = (initiativeId: string) =>
  queryOptions({
    queryKey: queryKeyInitiatives.updates(initiativeId),
    queryFn: async () => {
      const res = await getGoClient().initiatives.updates.list({ initiative_id: initiativeId });
      return { updates: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export async function createInitiative(input: TCreateInitiativeInput) {
  const res = await getGoClient().initiatives.create(input);
  return { data: res.data };
}

export async function updateInitiative(input: TUpdateInitiativeInput) {
  const res = await getGoClient().initiatives.update(input);
  return { data: res.data };
}

export type TInitiativeLifecycleAction = "archive" | "unarchive" | "delete";

export async function changeInitiativeLifecycle(input: {
  initiativeId: string;
  action: TInitiativeLifecycleAction;
}) {
  const initiatives = getGoClient().initiatives;
  const params = { initiative_id: input.initiativeId };
  switch (input.action) {
    case "archive":
      return { data: (await initiatives.archive(params)).data };
    case "unarchive":
      return { data: (await initiatives.unarchive(params)).data };
    case "delete":
      await initiatives.delete(params);
      return { data: null };
  }
}

export async function addInitiativeProject(input: { initiativeId: string; projectId: string }) {
  const res = await getGoClient().initiatives.projects.add({
    initiative_id: input.initiativeId,
    project_id: input.projectId,
  });
  return { data: res.data };
}

export async function removeInitiativeProject(input: { initiativeId: string; projectId: string }) {
  const res = await getGoClient().initiatives.projects.remove({
    initiative_id: input.initiativeId,
    project_id: input.projectId,
  });
  return { data: res.data };
}

export async function createInitiativeUpdate(input: TCreateInitiativeUpdateInput) {
  const res = await getGoClient().initiatives.updates.create(input);
  return { data: res.data };
}

export async function deleteInitiativeUpdate(input: { updateId: string }) {
  const res = await getGoClient().initiatives.updates.delete({ update_id: input.updateId });
  return { data: res.data };
}

function useInvalidateInitiatives() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyInitiatives.all() });
}

function useInvalidateInitiativesAndProjects() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyInitiatives.all() });
    await queryClient.invalidateQueries({ queryKey: ["projects"] });
  };
}

export function useCreateInitiativeMutation() {
  const invalidate = useInvalidateInitiativesAndProjects();
  return useMutation({
    mutationFn: createInitiative,
    meta: { errorTitle: "Could not create initiative" },
    onSuccess: invalidate,
  });
}

export function useUpdateInitiativeMutation() {
  const invalidate = useInvalidateInitiativesAndProjects();
  return useMutation({
    mutationFn: updateInitiative,
    meta: { errorTitle: "Could not update initiative" },
    onSuccess: invalidate,
  });
}

export function useInitiativeLifecycleMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: changeInitiativeLifecycle,
    meta: { errorTitle: "Could not change the initiative" },
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({
        queryKey:
          variables.action === "delete" ? queryKeyInitiatives.lists() : queryKeyInitiatives.all(),
      });
      await queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useAddInitiativeProjectMutation() {
  const invalidate = useInvalidateInitiativesAndProjects();
  return useMutation({
    mutationFn: addInitiativeProject,
    meta: { errorTitle: "Could not add project" },
    onSuccess: invalidate,
  });
}

export function useRemoveInitiativeProjectMutation() {
  const invalidate = useInvalidateInitiativesAndProjects();
  return useMutation({
    mutationFn: removeInitiativeProject,
    meta: { errorTitle: "Could not remove project" },
    onSuccess: invalidate,
  });
}

export function useCreateInitiativeUpdateMutation() {
  const invalidate = useInvalidateInitiatives();
  return useMutation({
    mutationFn: createInitiativeUpdate,
    meta: { errorTitle: "Could not post update" },
    onSuccess: invalidate,
  });
}

export function useDeleteInitiativeUpdateMutation() {
  const invalidate = useInvalidateInitiatives();
  return useMutation({
    mutationFn: deleteInitiativeUpdate,
    meta: { errorTitle: "Could not delete update" },
    onSuccess: invalidate,
  });
}

export const initiativeNameMaxLength = 80;

export const InitiativeFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(initiativeNameMaxLength, `Name should be at most ${initiativeNameMaxLength} characters.`),
  description: z.string().trim(),
  status: z.enum(["planned", "active", "completed"]),
  owner_id: z.string(),
  target_date: z.string(),
  icon: z.string(),
  color: z.string(),
});
