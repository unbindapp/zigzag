import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem } from "@/lib/server/types";

export type TFavorite = TListItem<TGoClient["favorites"]["list"]>;
export type TFavoriteKind = TFavorite["kind"];
export type TFavoriteTarget = TInput<TGoClient["favorites"]["add"]>;

export const queryKeyFavorites = {
  all: () => ["favorites"] as const,
  list: (workspaceId: string) => ["favorites", "list", workspaceId] as const,
};

export const favoritesListQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyFavorites.list(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().favorites.list({ workspace_id: workspaceId });
      return { favorites: res.data ?? [] };
    },
  });

export async function addFavorite(input: TFavoriteTarget) {
  const res = await getGoClient().favorites.add(input);
  return { data: res.data };
}

export async function removeFavorite(input: TFavoriteTarget) {
  const res = await getGoClient().favorites.remove(input);
  return { data: res.data };
}

export async function reorderFavorites(input: { workspaceId: string; favoriteIds: string[] }) {
  const res = await getGoClient().favorites.reorder({
    workspace_id: input.workspaceId,
    favorite_ids: input.favoriteIds,
  });
  return { data: res.data ?? [] };
}

function useInvalidateFavorites() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyFavorites.all() });
}

export function useAddFavoriteMutation() {
  const invalidate = useInvalidateFavorites();
  return useMutation({
    mutationFn: addFavorite,
    meta: { errorTitle: "Could not add favorite" },
    onSuccess: invalidate,
  });
}

export function useRemoveFavoriteMutation() {
  const invalidate = useInvalidateFavorites();
  return useMutation({
    mutationFn: removeFavorite,
    meta: { errorTitle: "Could not remove favorite" },
    onSuccess: invalidate,
  });
}

export function useReorderFavoritesMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderFavorites,
    meta: { errorTitle: "Could not reorder favorites" },
    onSuccess: (result, input) => {
      queryClient.setQueryData(queryKeyFavorites.list(input.workspaceId), {
        favorites: result.data,
      });
    },
  });
}
