import { queryOptions } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TResponseData } from "@/lib/server/types";

export type TSearchResults = TResponseData<TGoClient["search"]>;

const SEARCH_LIMIT = 8;
const SEARCH_STALE_MS = 30 * 1000;

export const queryKeySearch = {
  all: () => ["search"] as const,
  query: (workspaceId: string, query: string) => ["search", workspaceId, query] as const,
};

export const searchQuery = (workspaceId: string, query: string) =>
  queryOptions({
    queryKey: queryKeySearch.query(workspaceId, query),
    queryFn: async () => {
      const res = await getGoClient().search({
        workspace_id: workspaceId,
        query,
        limit: SEARCH_LIMIT,
      });
      return { results: res.data };
    },
    enabled: query.trim().length > 0,
    staleTime: SEARCH_STALE_MS,
  });
