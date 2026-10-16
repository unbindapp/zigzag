import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TResponseData } from "@/lib/server/types";

export type TSystemSettings = TResponseData<TGoClient["system"]["settings"]["get"]>;
export type TUpdateSystemSettingsInput = TInput<TGoClient["system"]["settings"]["put"]>;

export const queryKeySystem = {
  settings: () => ["system", "settings"] as const,
};

export const systemSettingsQuery = () =>
  queryOptions({
    queryKey: queryKeySystem.settings(),
    queryFn: async () => {
      const res = await getGoClient().system.settings.get();
      return { settings: res.data };
    },
  });

export async function updateSystemSettings(input: TUpdateSystemSettingsInput) {
  const res = await getGoClient().system.settings.put(input);
  return { data: res.data };
}

export function useUpdateSystemSettingsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateSystemSettings,
    meta: { errorTitle: "Could not update settings" },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeySystem.settings() }),
  });
}
