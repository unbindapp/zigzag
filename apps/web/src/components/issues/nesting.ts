type TNestable = { id: string; parent?: { id: string } | null };

export type TNestedIssue<T> = { issue: T; depth: number };

export function nestSubIssues<T extends TNestable>(issues: T[]): TNestedIssue<T>[] {
  const ids = new Set(issues.map((issue) => issue.id));
  const childrenOf = new Map<string, T[]>();
  const roots: T[] = [];
  for (const issue of issues) {
    const parentId = issue.parent?.id;
    if (!parentId || !ids.has(parentId) || parentId === issue.id) {
      roots.push(issue);
      continue;
    }
    childrenOf.set(parentId, [...(childrenOf.get(parentId) ?? []), issue]);
  }

  const result: TNestedIssue<T>[] = [];
  const visited = new Set<string>();
  const visit = (issue: T, depth: number) => {
    if (visited.has(issue.id)) return;
    visited.add(issue.id);
    result.push({ issue, depth });
    for (const child of childrenOf.get(issue.id) ?? []) visit(child, depth + 1);
  };
  roots.forEach((issue) => visit(issue, 0));
  issues.forEach((issue) => visit(issue, 0));
  return result;
}
