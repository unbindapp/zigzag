import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TImport = TListItem<TGoClient["imports"]["list"]>;
export type TImportReport = TResponseData<TGoClient["imports"]["linear"]>;
export type TImportLinearInput = TInput<TGoClient["imports"]["linear"]>;
export type TLinearExport = TImportLinearInput["export"];

const LIST_REFETCH_INTERVAL_MS = 15 * 1000;

export const queryKeyImports = {
  all: () => ["imports"] as const,
  list: (workspaceId: string) => ["imports", "list", workspaceId] as const,
};

export const importsListQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyImports.list(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().imports.list({ workspace_id: workspaceId });
      return { imports: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export async function importLinear(input: TImportLinearInput) {
  const res = await getGoClient().imports.linear(input);
  return { data: res.data };
}

export function useImportLinearMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: importLinear,
    meta: { errorTitle: "Could not import the Linear export" },
    onSuccess: ({ data }) => {
      if (data.dry_run) {
        return queryClient.invalidateQueries({ queryKey: queryKeyImports.all() });
      }
      return queryClient.invalidateQueries();
    },
  });
}
