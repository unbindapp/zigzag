import { useMutation, useQueryClient } from "@tanstack/react-query";

import { meQuery } from "@/lib/queries/me";
import { getGoClient } from "@/lib/server/client";

export async function login(input: { email: string; password: string }) {
  const res = await getGoClient().auth.login(input);
  return { data: res.data };
}

export async function logout() {
  const res = await getGoClient().auth.logout();
  return { data: res.data };
}

export async function requestPasswordReset(input: { email: string }) {
  const res = await getGoClient().auth.requestPasswordReset(input);
  return { data: res.data };
}

export async function resetPassword(input: { token: string; password: string }) {
  const res = await getGoClient().auth.resetPassword(input);
  return { data: res.data };
}

export function useLogoutMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: logout,
    meta: { errorTitle: "Could not sign out" },
    onSuccess: () => {
      queryClient.clear();
      queryClient.setQueryData(meQuery.queryKey, null);
    },
  });
}
