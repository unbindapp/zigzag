import { queryOptions } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput } from "@/lib/server/types";

export type TBootstrapInput = TInput<TGoClient["setup"]["createUser"]>;

export const queryKeySetup = {
  status: () => ["setup", "status"] as const,
};

export const setupStatusQuery = () =>
  queryOptions({
    queryKey: queryKeySetup.status(),
    queryFn: async () => {
      const res = await getGoClient().setup.status();
      return { status: res.data };
    },
    staleTime: 0,
  });

export async function createBootstrapUser(input: TBootstrapInput) {
  const res = await getGoClient().setup.createUser(input);
  return { data: res.data };
}
