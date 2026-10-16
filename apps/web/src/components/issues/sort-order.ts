const SORT_ORDER_STEP = 1;

export function sortOrderBetween(before: number | undefined, after: number | undefined) {
  if (before === undefined && after === undefined) return 0;
  if (before === undefined) return (after as number) - SORT_ORDER_STEP;
  if (after === undefined) return before + SORT_ORDER_STEP;
  return (before + after) / 2;
}

export function sortOrderAtIndex(neighbors: number[], index: number) {
  const clamped = Math.max(0, Math.min(index, neighbors.length));
  return sortOrderBetween(neighbors[clamped - 1], neighbors[clamped]);
}
