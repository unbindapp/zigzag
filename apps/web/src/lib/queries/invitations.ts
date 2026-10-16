import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { queryKeyMe } from "@/lib/queries/me";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TResponseData } from "@/lib/server/types";

export type TInvitationPreview = TResponseData<TGoClient["invitations"]["get"]>;
export type TAcceptInvitationInput = TInput<TGoClient["invitations"]["accept"]>;

export const queryKeyInvitations = {
  preview: (token: string) => ["invitations", "preview", token] as const,
};

export const invitationPreviewQuery = (token: string) =>
  queryOptions({
    queryKey: queryKeyInvitations.preview(token),
    queryFn: async () => {
      const res = await getGoClient().invitations.get({ token });
      return { invitation: res.data };
    },
    retry: false,
  });

export async function acceptInvitation(input: TAcceptInvitationInput) {
  const res = await getGoClient().invitations.accept(input);
  return { data: res.data };
}

export function useAcceptInvitationMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: acceptInvitation,
    meta: { errorTitle: "Could not accept invitation" },
    onSuccess: async () => {
      queryClient.removeQueries({ queryKey: queryKeyMe });
      await queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}
