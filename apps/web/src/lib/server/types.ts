import type { getGoClient } from "@/lib/server/client";

export type TGoClient = ReturnType<typeof getGoClient>;

type TEndpoint = (...args: never[]) => Promise<{ data: unknown }>;

export type TResponseData<F extends TEndpoint> = Awaited<ReturnType<F>>["data"];
export type TInput<F extends TEndpoint> = Parameters<F>[0];
export type TListItem<F extends TEndpoint> =
  NonNullable<TResponseData<F>> extends (infer T)[] ? T : never;
