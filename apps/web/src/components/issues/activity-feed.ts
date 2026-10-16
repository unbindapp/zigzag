export type TFeedStream<T extends { at: string }> = { items: T[]; hasMore: boolean };

export function mergeFeed<T extends { at: string }>(streams: TFeedStream<T>[]) {
  const cutoffs = streams
    .filter((stream) => stream.hasMore && stream.items.length > 0)
    .map((stream) => Date.parse(stream.items[stream.items.length - 1].at));
  const cutoff = cutoffs.length > 0 ? Math.min(...cutoffs) : Infinity;
  return streams
    .flatMap((stream) => stream.items)
    .filter((item) => Date.parse(item.at) <= cutoff)
    .sort((a, b) => Date.parse(a.at) - Date.parse(b.at));
}
