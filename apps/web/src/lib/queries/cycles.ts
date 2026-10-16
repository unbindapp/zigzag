import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { LIST_REFETCH_INTERVAL_MS } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";
import { isApiError } from "@/lib/server/errors";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TCycle = TListItem<TGoClient["cycles"]["list"]>;
export type TCycleDetail = TResponseData<TGoClient["cycles"]["get"]>;
export type TCycleScopePoint = NonNullable<TCycleDetail["scope_history"]>[number];
export type TCreateCycleInput = TInput<TGoClient["cycles"]["create"]>;
export type TUpdateCycleInput = TInput<TGoClient["cycles"]["update"]>;

export const queryKeyCycles = {
  all: () => ["cycles"] as const,
  list: (teamId: string) => ["cycles", "list", teamId] as const,
  detail: (teamId: string, number: number) => ["cycles", "detail", teamId, number] as const,
};

export const cyclesListQuery = (teamId: string) =>
  queryOptions({
    queryKey: queryKeyCycles.list(teamId),
    queryFn: async () => {
      const res = await getGoClient().cycles.list({ team_id: teamId });
      return { cycles: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const cycleByNumberQuery = (teamId: string, number: number) =>
  queryOptions({
    queryKey: queryKeyCycles.detail(teamId, number),
    queryFn: async () => {
      const res = await getGoClient().cycles.get({ team_id: teamId, number });
      return { cycle: res.data };
    },
    retry: retryUnlessNotFound,
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export async function createCycle(input: TCreateCycleInput) {
  const res = await getGoClient().cycles.create(input);
  return { data: res.data };
}

export async function updateCycle(input: TUpdateCycleInput) {
  const res = await getGoClient().cycles.update(input);
  return { data: res.data };
}

export async function archiveCycle(input: { cycleId: string }) {
  const res = await getGoClient().cycles.archive({ cycle_id: input.cycleId });
  return { data: res.data };
}

export async function deleteCycle(input: { cycleId: string }) {
  const res = await getGoClient().cycles.delete({ cycle_id: input.cycleId });
  return { data: res.data };
}

function useInvalidateCycles() {
  const queryClient = useQueryClient();
  return async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeyCycles.all() });
    await queryClient.invalidateQueries({ queryKey: ["teams"] });
    await queryClient.invalidateQueries({ queryKey: ["favorites"] });
  };
}

export function useCreateCycleMutation() {
  const invalidate = useInvalidateCycles();
  return useMutation({
    mutationFn: createCycle,
    meta: { errorTitle: "Could not create cycle" },
    onSuccess: invalidate,
  });
}

export function useUpdateCycleMutation() {
  const invalidate = useInvalidateCycles();
  return useMutation({
    mutationFn: updateCycle,
    meta: { errorTitle: "Could not update cycle" },
    onSuccess: invalidate,
  });
}

export function useArchiveCycleMutation() {
  const invalidate = useInvalidateCycles();
  return useMutation({
    mutationFn: archiveCycle,
    meta: { errorTitle: "Could not archive cycle" },
    onSuccess: invalidate,
  });
}

export function useDeleteCycleMutation() {
  const invalidate = useInvalidateCycles();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteCycle,
    meta: { errorTitle: "Could not delete cycle" },
    onSuccess: async () => {
      await invalidate();
      await queryClient.invalidateQueries({ queryKey: ["issues"] });
    },
  });
}

export function cycleName(cycle: { name: string; number: number }) {
  return cycle.name || `Cycle ${cycle.number}`;
}

function retryUnlessNotFound(failureCount: number, error: Error) {
  return !isApiError(error, 404) && failureCount < 3;
}
