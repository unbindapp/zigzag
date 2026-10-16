import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TWebhook = TListItem<TGoClient["webhooks"]["list"]>;
export type TWebhookCreated = TResponseData<TGoClient["webhooks"]["create"]>;
export type TCreateWebhookInput = TInput<TGoClient["webhooks"]["create"]>;
export type TUpdateWebhookInput = TInput<TGoClient["webhooks"]["update"]>;
export const queryKeyWebhooks = {
  all: () => ["webhooks"] as const,
  list: (workspaceId: string) => ["webhooks", "list", workspaceId] as const,
};

export const webhooksListQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyWebhooks.list(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().webhooks.list({ workspace_id: workspaceId });
      return { webhooks: res.data ?? [] };
    },
  });

export async function createWebhook(input: TCreateWebhookInput) {
  const res = await getGoClient().webhooks.create(input);
  return { data: res.data };
}

export async function updateWebhook(input: TUpdateWebhookInput) {
  const res = await getGoClient().webhooks.update(input);
  return { data: res.data };
}

export async function deleteWebhook({ webhookId }: { webhookId: string }) {
  const res = await getGoClient().webhooks.delete({ webhook_id: webhookId });
  return { data: res.data };
}

export async function testWebhook({ webhookId }: { webhookId: string }) {
  const res = await getGoClient().webhooks.test({ webhook_id: webhookId });
  return { data: res.data };
}

function useInvalidateWebhooks() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyWebhooks.all() });
}

export function useCreateWebhookMutation() {
  const invalidate = useInvalidateWebhooks();
  return useMutation({
    mutationFn: createWebhook,
    meta: { errorTitle: "Could not create webhook" },
    onSuccess: invalidate,
  });
}

export function useUpdateWebhookMutation() {
  const invalidate = useInvalidateWebhooks();
  return useMutation({
    mutationFn: updateWebhook,
    meta: { errorTitle: "Could not update webhook" },
    onSuccess: invalidate,
  });
}

export function useDeleteWebhookMutation() {
  const invalidate = useInvalidateWebhooks();
  return useMutation({
    mutationFn: deleteWebhook,
    meta: { errorTitle: "Could not delete webhook" },
    onSuccess: invalidate,
  });
}

export function useTestWebhookMutation() {
  return useMutation({
    mutationFn: testWebhook,
    meta: { errorTitle: "Could not send a test event" },
  });
}
