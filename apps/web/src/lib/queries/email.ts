import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { queryKeySystem } from "@/lib/queries/system";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TResponseData } from "@/lib/server/types";

export type TEmailSettings = TResponseData<TGoClient["system"]["email"]["get"]>;
export type TEmailSettingsUpdateInput = TInput<TGoClient["system"]["email"]["put"]>;
export type TEmailEncryption = TEmailSettingsUpdateInput["encryption"];

export const queryKeyEmail = {
  settings: () => ["system", "email"] as const,
};

export const emailSettingsQuery = () =>
  queryOptions({
    queryKey: queryKeyEmail.settings(),
    queryFn: async () => {
      const res = await getGoClient().system.email.get();
      return { data: res.data };
    },
  });

export async function updateEmailSettings(input: TEmailSettingsUpdateInput) {
  const res = await getGoClient().system.email.put(input);
  return { data: res.data };
}

export async function clearEmailSettings() {
  const res = await getGoClient().system.email.delete();
  return { data: res.data };
}

export async function sendTestEmail() {
  const res = await getGoClient().system.email.test();
  return { data: res.data };
}

function useInvalidateEmail() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyEmail.settings() });
    await queryClient.invalidateQueries({ queryKey: queryKeySystem.settings() });
  };
}

export function useUpdateEmailSettingsMutation() {
  const invalidate = useInvalidateEmail();
  return useMutation({
    mutationFn: updateEmailSettings,
    meta: { errorTitle: "Could not save email settings" },
    onSuccess: invalidate,
  });
}

export function useClearEmailSettingsMutation() {
  const invalidate = useInvalidateEmail();
  return useMutation({
    mutationFn: clearEmailSettings,
    meta: { errorTitle: "Could not remove email settings" },
    onSuccess: invalidate,
  });
}

export function useSendTestEmailMutation() {
  return useMutation({
    mutationFn: sendTestEmail,
    meta: { errorTitle: "Test email failed" },
  });
}
