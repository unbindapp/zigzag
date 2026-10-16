import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TApiKeyShallow = TListItem<TGoClient["apiKeys"]["list"]>;
export type TApiKeyCreated = TResponseData<TGoClient["apiKeys"]["create"]>;
export type TApiKeyCreateInput = TInput<TGoClient["apiKeys"]["create"]>;

export const queryKeyApiKeys = {
  list: () => ["api-keys", "list"] as const,
};

export const apiKeysListQuery = () =>
  queryOptions({
    queryKey: queryKeyApiKeys.list(),
    queryFn: async () => {
      const res = await getGoClient().apiKeys.list({});
      return { apiKeys: res.data ?? [] };
    },
  });

export async function createApiKey(input: TApiKeyCreateInput) {
  const res = await getGoClient().apiKeys.create(input);
  return { data: res.data };
}

export async function deleteApiKey(input: { id: string }) {
  const res = await getGoClient().apiKeys.delete({ id: input.id });
  return { data: res.data };
}

export function useCreateApiKeyMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createApiKey,
    meta: { errorTitle: "Could not create API key" },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeyApiKeys.list() }),
  });
}

export function useDeleteApiKeyMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteApiKey,
    meta: { errorTitle: "Could not revoke API key" },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeyApiKeys.list() }),
  });
}
