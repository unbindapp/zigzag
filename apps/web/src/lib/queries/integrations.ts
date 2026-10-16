import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TGithubStatus = TResponseData<TGoClient["integrations"]["github"]["get"]>;
export type TGithubIntegration = NonNullable<TGithubStatus["integration"]>;
export type TGithubRepository = TGithubIntegration["repositories"][number];
export type TGithubSecret = TResponseData<TGoClient["integrations"]["github"]["create"]>;
export type TGithubRepoSync = TListItem<TGoClient["integrations"]["github"]["repoSyncs"]["list"]>;
export type TAddGithubRepoSyncInput = TInput<
  TGoClient["integrations"]["github"]["repoSyncs"]["add"]
>;

export const GITHUB_POLL_MS = 2000;

export const queryKeyIntegrations = {
  all: () => ["integrations"] as const,
  github: (workspaceId: string) => ["integrations", "github", workspaceId] as const,
  repoSyncs: (workspaceId: string) =>
    ["integrations", "github", workspaceId, "repo-syncs"] as const,
  repoSync: (repoSyncId: string) => ["integrations", "github", "repo-sync", repoSyncId] as const,
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

export async function updateGithubToken({
  workspaceId,
  token,
}: {
  workspaceId: string;
  token: string;
}) {
  const res = await getGoClient().integrations.github.update({ workspace_id: workspaceId, token });
  return { data: res.data };
}

export function useUpdateGithubTokenMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: updateGithubToken,
    meta: { errorTitle: "Could not save the GitHub token" },
    onSuccess: invalidate,
  });
}

export const githubRepoSyncsQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyIntegrations.repoSyncs(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().integrations.github.repoSyncs.list({
        workspace_id: workspaceId,
      });
      return { repoSyncs: res.data ?? [] };
    },
  });

export const githubRepoSyncQuery = (repoSyncId: string) =>
  queryOptions({
    queryKey: queryKeyIntegrations.repoSync(repoSyncId),
    queryFn: async () => {
      const res = await getGoClient().integrations.github.repoSyncs.get({
        repo_sync_id: repoSyncId,
      });
      return { repoSync: res.data };
    },
  });

export async function addGithubRepoSync(input: TAddGithubRepoSyncInput) {
  const res = await getGoClient().integrations.github.repoSyncs.add(input);
  return { data: res.data };
}

export async function backfillGithubRepoSync({ repoSyncId }: { repoSyncId: string }) {
  const res = await getGoClient().integrations.github.repoSyncs.backfill({
    repo_sync_id: repoSyncId,
  });
  return { data: res.data };
}

export async function removeGithubRepoSync({ repoSyncId }: { repoSyncId: string }) {
  const res = await getGoClient().integrations.github.repoSyncs.remove({
    repo_sync_id: repoSyncId,
  });
  return { data: res.data };
}

export function useAddGithubRepoSyncMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: addGithubRepoSync,
    meta: { errorTitle: "Could not add the repository" },
    onSuccess: invalidate,
  });
}

export function useBackfillGithubRepoSyncMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: backfillGithubRepoSync,
    meta: { errorTitle: "Could not import existing issues" },
    onSuccess: invalidate,
  });
}

export function useRemoveGithubRepoSyncMutation() {
  const invalidate = useInvalidateIntegrations();
  return useMutation({
    mutationFn: removeGithubRepoSync,
    meta: { errorTitle: "Could not remove the repository" },
    onSuccess: invalidate,
  });
}
