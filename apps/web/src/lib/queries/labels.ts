import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { hexColorPattern } from "@/lib/helpers/colors";
import { LIST_REFETCH_INTERVAL_MS } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TLabelTree = TListItem<TGoClient["labels"]["list"]>;
export type TLabel = TResponseData<TGoClient["labels"]["create"]>;
export type TCreateLabelInput = TInput<TGoClient["labels"]["create"]>;
export type TUpdateLabelInput = TInput<TGoClient["labels"]["update"]>;

export type TLabelScope = {
  workspaceId: string;
  teamId?: string;
  allTeams?: boolean;
  includeArchived?: boolean;
};

export const queryKeyLabels = {
  all: () => ["labels"] as const,
  tree: (scope: TLabelScope) =>
    [
      "labels",
      "tree",
      scope.workspaceId,
      scope.teamId ?? null,
      Boolean(scope.allTeams),
      Boolean(scope.includeArchived),
    ] as const,
};

export const labelTreeQuery = (scope: TLabelScope) =>
  queryOptions({
    queryKey: queryKeyLabels.tree(scope),
    queryFn: async () => {
      const res = await getGoClient().labels.list({
        workspace_id: scope.workspaceId,
        team_id: scope.teamId,
        all_teams: scope.allTeams,
        include_archived: scope.includeArchived,
      });
      return { labels: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export async function createLabel(input: TCreateLabelInput) {
  const res = await getGoClient().labels.create(input);
  return { data: res.data };
}

export async function updateLabel(input: TUpdateLabelInput) {
  const res = await getGoClient().labels.update(input);
  return { data: res.data };
}

export async function archiveLabel(input: { labelId: string }) {
  const res = await getGoClient().labels.archive({ label_id: input.labelId });
  return { data: res.data };
}

export async function unarchiveLabel(input: { labelId: string }) {
  const res = await getGoClient().labels.unarchive({ label_id: input.labelId });
  return { data: res.data };
}

export async function deleteLabel(input: { labelId: string }) {
  const res = await getGoClient().labels.delete({ label_id: input.labelId });
  return { data: res.data };
}

function useInvalidateLabels() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyLabels.all() });
    await queryClient.invalidateQueries({ queryKey: ["favorites"] });
  };
}

export function useCreateLabelMutation() {
  const invalidate = useInvalidateLabels();
  return useMutation({
    mutationFn: createLabel,
    meta: { errorTitle: "Could not create label" },
    onSuccess: invalidate,
  });
}

export function useUpdateLabelMutation() {
  const invalidate = useInvalidateLabels();
  return useMutation({
    mutationFn: updateLabel,
    meta: { errorTitle: "Could not update label" },
    onSuccess: invalidate,
  });
}

export function useArchiveLabelMutation() {
  const invalidate = useInvalidateLabels();
  return useMutation({
    mutationFn: archiveLabel,
    meta: { errorTitle: "Could not archive label" },
    onSuccess: invalidate,
  });
}

export function useUnarchiveLabelMutation() {
  const invalidate = useInvalidateLabels();
  return useMutation({
    mutationFn: unarchiveLabel,
    meta: { errorTitle: "Could not restore label" },
    onSuccess: invalidate,
  });
}

export function useDeleteLabelMutation() {
  const invalidate = useInvalidateLabels();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteLabel,
    meta: { errorTitle: "Could not delete label" },
    onSuccess: async () => {
      await invalidate();
      await queryClient.invalidateQueries({ queryKey: ["issues"] });
    },
  });
}

export const labelNameMaxLength = 64;

export const LabelFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(labelNameMaxLength, `Name should be at most ${labelNameMaxLength} characters.`),
  color: z.string().regex(hexColorPattern, "Pick a color."),
  description: z.string().trim().max(500, "Keep the description under 500 characters."),
  parentId: z.string(),
  groupSingleSelect: z.boolean(),
});
