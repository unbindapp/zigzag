import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TResponseData } from "@/lib/server/types";

export type TGoogleAuthSettings = TResponseData<TGoClient["system"]["googleAuth"]["get"]>;
export type TGoogleAuthSettingsUpdateInput = TInput<TGoClient["system"]["googleAuth"]["put"]>;

export const queryKeyGoogleAuth = {
  settings: () => ["system", "google-auth"] as const,
  providers: () => ["auth", "providers"] as const,
};

export const googleAuthSettingsQuery = () =>
  queryOptions({
    queryKey: queryKeyGoogleAuth.settings(),
    queryFn: async () => {
      const res = await getGoClient().system.googleAuth.get();
      return { data: res.data };
    },
  });

export const authProvidersQuery = () =>
  queryOptions({
    queryKey: queryKeyGoogleAuth.providers(),
    queryFn: async () => {
      const res = await getGoClient().auth.providers();
      return { providers: res.data };
    },
  });

export function googleSignInHref(redirect: string | undefined) {
  const query = redirect ? `?${new URLSearchParams({ redirect })}` : "";
  return `/api/go/auth/google/start${query}`;
}

export async function updateGoogleAuthSettings(input: TGoogleAuthSettingsUpdateInput) {
  const res = await getGoClient().system.googleAuth.put(input);
  return { data: res.data };
}

export async function clearGoogleAuthSettings() {
  const res = await getGoClient().system.googleAuth.delete();
  return { data: res.data };
}

function useInvalidateGoogleAuth() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyGoogleAuth.settings() });
    await queryClient.invalidateQueries({ queryKey: queryKeyGoogleAuth.providers() });
  };
}

export function useUpdateGoogleAuthSettingsMutation() {
  const invalidate = useInvalidateGoogleAuth();
  return useMutation({
    mutationFn: updateGoogleAuthSettings,
    meta: { errorTitle: "Could not save Google sign-in settings" },
    onSuccess: invalidate,
  });
}

export function useClearGoogleAuthSettingsMutation() {
  const invalidate = useInvalidateGoogleAuth();
  return useMutation({
    mutationFn: clearGoogleAuthSettings,
    meta: { errorTitle: "Could not remove Google sign-in settings" },
    onSuccess: invalidate,
  });
}
