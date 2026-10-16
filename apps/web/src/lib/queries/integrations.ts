import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TResponseData } from "@/lib/server/types";

export type TGithubStatus = TResponseData<TGoClient["integrations"]["github"]["get"]>;
export type TGithubIntegration = NonNullable<TGithubStatus["integration"]>;
export type TGithubRepository = TGithubIntegration["repositories"][number];
export type TGithubSecret = TResponseData<TGoClient["integrations"]["github"]["create"]>;

export const queryKeyIntegrations = {
  all: () => ["integrations"] as const,
  github: (workspaceId: string) => ["integrations", "github", workspaceId] as const,
};

export const githubIntegrationQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyIntegrations.github(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().integrations.github.get({ workspace_id: workspaceId });
      return { status: res.data };
    },
  });

export async function createGithubIntegration({ workspaceId }: { workspaceId: string }) {
  const res = await getGoClient().integrations.github.create({ workspace_id: workspaceId });
  return { data: res.data };
}

export async function regenerateGithubSecret({ workspaceId }: { workspaceId: string }) {
  const res = await getGoClient().integrations.github.regenerateSecret({
    workspace_id: workspaceId,
  });
  return { data: res.data };
}

export async function deleteGithubIntegration({ workspaceId }: { workspaceId: string }) {
  const res = await getGoClient().integrations.github.delete({ workspace_id: workspaceId });
  return { data: res.data };
}

function useInvalidateIntegrations() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyIntegrations.all() });
}

export function useCreateGithubIntegrationMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: createGithubIntegration,
    meta: { errorTitle: "Could not connect GitHub" },
    onSuccess: invalidate,
  });
}

export function useRegenerateGithubSecretMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: regenerateGithubSecret,
    meta: { errorTitle: "Could not regenerate the secret" },
    onSuccess: invalidate,
  });
}

export function useDeleteGithubIntegrationMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: deleteGithubIntegration,
    meta: { errorTitle: "Could not disconnect GitHub" },
    onSuccess: invalidate,
  });
}
