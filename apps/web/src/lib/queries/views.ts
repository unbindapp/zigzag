import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { LIST_REFETCH_INTERVAL_MS } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";
import { isApiError } from "@/lib/server/errors";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TView = TListItem<TGoClient["views"]["list"]>;
export type TViewDetail = TResponseData<TGoClient["views"]["get"]>;
export type TViewDisplay = TView["display"];
export type TCreateViewInput = TInput<TGoClient["views"]["create"]>;
export type TUpdateViewInput = TInput<TGoClient["views"]["update"]>;

export const queryKeyViews = {
  all: () => ["views"] as const,
  list: (workspaceId: string) => ["views", "list", workspaceId] as const,
  detail: (viewId: string) => ["views", "detail", viewId] as const,
};

export const viewsListQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyViews.list(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().views.list({ workspace_id: workspaceId, kind: "issue" });
      return { views: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const viewQuery = (viewId: string) =>
  queryOptions({
    queryKey: queryKeyViews.detail(viewId),
    queryFn: async () => {
      const res = await getGoClient().views.get({ view_id: viewId });
      return { view: res.data };
    },
    retry: retryUnlessNotFound,
  });

export async function createView(input: TCreateViewInput) {
  const res = await getGoClient().views.create(input);
  return { data: res.data };
}

export async function updateView(input: TUpdateViewInput) {
  const res = await getGoClient().views.update(input);
  return { data: res.data };
}

export async function deleteView(input: { viewId: string }) {
  const res = await getGoClient().views.delete({ view_id: input.viewId });
  return { data: res.data };
}

function useInvalidateViews() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyViews.all() });
    await queryClient.invalidateQueries({ queryKey: ["favorites"] });
  };
}

export function useCreateViewMutation() {
  const invalidate = useInvalidateViews();
  return useMutation({
    mutationFn: createView,
    meta: { errorTitle: "Could not save view" },
    onSuccess: invalidate,
  });
}

export function useUpdateViewMutation() {
  const invalidate = useInvalidateViews();
  return useMutation({
    mutationFn: updateView,
    meta: { errorTitle: "Could not update view" },
    onSuccess: invalidate,
  });
}

export function useDeleteViewMutation() {
  const invalidate = useInvalidateViews();
  return useMutation({
    mutationFn: deleteView,
    meta: { errorTitle: "Could not delete view" },
    onSuccess: invalidate,
  });
}

export const viewNameMaxLength = 64;

export const ViewFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(viewNameMaxLength, `Name should be at most ${viewNameMaxLength} characters.`),
  description: z.string().trim().max(500, "Keep the description under 500 characters."),
  icon: z.string(),
  color: z.string(),
  shared: z.boolean(),
});

function retryUnlessNotFound(failureCount: number, error: Error) {
  return !isApiError(error, 404) && failureCount < 3;
}
