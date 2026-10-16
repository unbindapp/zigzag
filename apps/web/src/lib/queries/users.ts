import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem } from "@/lib/server/types";

export type TUser = TListItem<TGoClient["users"]["list"]>;
export type TCreateUserInput = TInput<TGoClient["users"]["create"]>;
export type TUpdateUserInput = TInput<TGoClient["users"]["update"]>;

export const queryKeyUsers = {
  all: () => ["users"] as const,
  list: (query: string) => ["users", "list", query] as const,
};

export const usersListQuery = (query = "") =>
  queryOptions({
    queryKey: queryKeyUsers.list(query),
    queryFn: async () => {
      const res = await getGoClient().users.list({ query: query || undefined, limit: 250 });
      return { users: res.data ?? [] };
    },
  });

export async function createUser(input: TCreateUserInput) {
  const res = await getGoClient().users.create(input);
  return { data: res.data };
}

export async function updateUser(input: TUpdateUserInput) {
  const res = await getGoClient().users.update(input);
  return { data: res.data };
}

export async function deleteUser(input: { userId: string }) {
  const res = await getGoClient().users.delete({ user_id: input.userId });
  return { data: res.data };
}

function useInvalidateUsers() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyUsers.all() });
}

export function useCreateUserMutation() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: createUser,
    meta: { errorTitle: "Could not create user" },
    onSuccess: invalidate,
  });
}

export function useUpdateUserMutation() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: updateUser,
    meta: { errorTitle: "Could not update user" },
    onSuccess: invalidate,
  });
}

export function useDeleteUserMutation() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: deleteUser,
    meta: { errorTitle: "Could not delete user" },
    onSuccess: invalidate,
  });
}
