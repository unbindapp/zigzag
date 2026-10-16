import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TConnectedApp = TListItem<TGoClient["connectedApps"]["list"]>;
export type TConnectedAppClient = TResponseData<TGoClient["connectedApps"]["client"]>;
export type TConnectedAppApproveInput = TInput<TGoClient["connectedApps"]["approve"]>;
export type TConnectedAppDenyInput = TInput<TGoClient["connectedApps"]["deny"]>;

export const queryKeyConnectedApps = {
  list: () => ["connected-apps", "list"] as const,
  client: (clientId: string, redirectUri: string) =>
    ["connected-apps", "client", clientId, redirectUri] as const,
};

export const connectedAppsListQuery = () =>
  queryOptions({
    queryKey: queryKeyConnectedApps.list(),
    queryFn: async () => {
      const res = await getGoClient().connectedApps.list();
      return { connectedApps: res.data ?? [] };
    },
  });

export const connectedAppClientQuery = ({
  clientId,
  redirectUri,
}: {
  clientId: string;
  redirectUri: string;
}) =>
  queryOptions({
    queryKey: queryKeyConnectedApps.client(clientId, redirectUri),
    queryFn: async () => {
      const res = await getGoClient().connectedApps.client({
        client_id: clientId,
        redirect_uri: redirectUri,
      });
      return { client: res.data };
    },
    staleTime: Infinity,
    retry: false,
  });

export async function approveConnectedApp(input: TConnectedAppApproveInput) {
  const res = await getGoClient().connectedApps.approve(input);
  return { data: res.data };
}

export async function denyConnectedApp(input: TConnectedAppDenyInput) {
  const res = await getGoClient().connectedApps.deny(input);
  return { data: res.data };
}

export async function revokeConnectedApp(input: { id: string }) {
  const res = await getGoClient().connectedApps.revoke({ id: input.id });
  return { data: res.data };
}

export function useRevokeConnectedAppMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: revokeConnectedApp,
    meta: { errorTitle: "Could not revoke access" },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeyConnectedApps.list() }),
  });
}

export function useApproveConnectedAppMutation() {
  return useMutation({
    mutationFn: approveConnectedApp,
    meta: { errorTitle: "Could not approve" },
  });
}

export function useDenyConnectedAppMutation() {
  return useMutation({
    mutationFn: denyConnectedApp,
    meta: { errorTitle: "Could not deny" },
  });
}
