import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import { isApiError } from "@/lib/server/errors";
import type { TGoClient, TInput, TResponseData } from "@/lib/server/types";

export type TMe = TResponseData<TGoClient["users"]["whoami"]>;
export type TMeWorkspace = TMe["workspaces"] extends (infer W)[] | null | undefined ? W : never;
export type TNotificationSettings = TResponseData<
  TGoClient["users"]["notificationSettings"]["get"]
>;
export type TUpdateMeInput = TInput<TGoClient["users"]["me"]>;

export const NO_AVATAR_UPLOAD_ID = "00000000-0000-0000-0000-000000000000";

export const queryKeyMe = ["me"] as const;

export const queryKeyMeGroup = {
  me: () => queryKeyMe,
  notificationSettings: () => ["me", "notification-settings"] as const,
};

export const meQuery = queryOptions({
  queryKey: queryKeyMe,
  queryFn: async (): Promise<TMe | null> => {
    try {
      const res = await getGoClient().users.whoami();
      return res.data;
    } catch (error) {
      if (isApiError(error, 401)) return null;
      throw error;
    }
  },
  staleTime: 5 * 60 * 1000,
});

export function isSystemAdmin(me: TMe | null | undefined) {
  return me?.system_admin ?? false;
}

export function meWorkspaces(me: TMe | null | undefined): TMeWorkspace[] {
  return me?.workspaces ?? [];
}

export const notificationSettingsQuery = () =>
  queryOptions({
    queryKey: queryKeyMeGroup.notificationSettings(),
    queryFn: async () => {
      const res = await getGoClient().users.notificationSettings.get();
      return { settings: res.data };
    },
  });

export async function updateMe(input: TUpdateMeInput) {
  const res = await getGoClient().users.me(input);
  return { data: res.data };
}

export async function changePassword(input: { currentPassword: string; newPassword: string }) {
  const res = await getGoClient().users.password({
    current_password: input.currentPassword,
    new_password: input.newPassword,
  });
  return { data: res.data };
}

export async function updateNotificationSettings(input: TNotificationSettings) {
  const res = await getGoClient().users.notificationSettings.put(input);
  return { data: res.data };
}

export function useUpdateMeMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateMe,
    meta: { errorTitle: "Could not update profile" },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeyMe });
      await queryClient.invalidateQueries({ queryKey: ["workspaces", "members"] });
    },
  });
}

export function useChangePasswordMutation() {
  return useMutation({
    mutationFn: changePassword,
    meta: { errorTitle: "Could not change password" },
  });
}

export function useUpdateNotificationSettingsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateNotificationSettings,
    meta: { errorTitle: "Could not update notifications" },
    onSuccess: (res) => {
      queryClient.setQueryData(queryKeyMeGroup.notificationSettings(), { settings: res.data });
    },
  });
}
