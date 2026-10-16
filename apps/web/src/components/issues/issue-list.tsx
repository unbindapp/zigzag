import { useVirtualizer } from "@tanstack/react-virtual";
import { ListTodoIcon, PlusIcon } from "lucide-react";
import {
  type MouseEvent,
  type ReactNode,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";

import { BulkActionBar } from "@/components/issues/bulk-action-bar";
import type { DisplaySettings } from "@/components/issues/display";
import type { ListIssuesFilters } from "@/components/issues/filters";
import type { TIssueGroup } from "@/components/issues/grouping";
import { IssueGroupHeader } from "@/components/issues/issue-group-header";
import { useIsPeekable, useIssuePeek } from "@/components/issues/issue-peek";
import { IssueRow } from "@/components/issues/issue-row";
import { nestSubIssues } from "@/components/issues/nesting";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import type { TIssue } from "@/lib/queries/issues";
import {
  newIssueDefaultsForGroup,
  useIssueCollection,
} from "@/components/issues/use-issue-collection";
import { useOpenIssuePage } from "@/components/issues/use-issue-actions";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { List } from "@/components/layout/list";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useHotkeyScope } from "@/lib/hotkeys/use-hotkey-scope";
import { useIssueSelectionStore } from "@/lib/stores/issue-selection";

export type TIssueListProps = {
  workspaceId: string;
  filters: ListIssuesFilters;
  display?: Partial<DisplaySettings>;
  fixedFilters?: Partial<ListIssuesFilters>;
  emptyState?: ReactNode;
};

const ROW_HEIGHT = 36;
const OVERSCAN = 12;
const PREFETCH_ROWS = 20;

type TListItem =
  | { kind: "header"; key: string; group: TIssueGroup }
  | { kind: "row"; key: string; issue: TIssue; depth: number };

function findScrollParent(element: HTMLElement | null): HTMLElement | null {
  let current = element?.parentElement ?? null;
  while (current) {
    const overflowY = getComputedStyle(current).overflowY;
    if (overflowY === "auto" || overflowY === "scroll") return current;
    current = current.parentElement;
  }
  return null;
}

export function IssueList(props: TIssueListProps) {
  const { query, issues, groups, display, fixedTeamId } = useIssueCollection(props);
  const { canEdit } = useWorkspace();
  const openNewIssue = useNewIssueStore((state) => state.open);

  if (query.isPending) return <ListSkeleton rows={8} />;
  if (query.error) {
    return <ErrorCard message={query.error.message} onRetry={() => void query.refetch()} />;
  }
  if (issues.length === 0) {
    return (
      props.emptyState ?? (
        <EmptyState
          icon={ListTodoIcon}
          title="No issues yet"
          description="Issues that match these filters show up here."
          action={
            canEdit && (
              <Button
                size="sm"
                variant="outline"
                onClick={() => openNewIssue(fixedTeamId ? { teamId: fixedTeamId } : undefined)}
              >
                <PlusIcon className="size-4" />
                New issue
              </Button>
            )
          }
        />
      )
    );
  }
  return (
    <VirtualIssueList
      groups={groups}
      issues={issues}
      showHeaders={display.group_by !== "none"}
      nestSubIssues={display.nest_sub_issues}
      properties={display.properties}
      fixedTeamId={fixedTeamId}
      hasNextPage={query.hasNextPage}
      isFetchingNextPage={query.isFetchingNextPage}
      fetchNextPage={() => void query.fetchNextPage()}
    />
  );
}

type TVirtualProps = {
  groups: TIssueGroup[];
  issues: TIssue[];
  showHeaders: boolean;
  nestSubIssues: boolean;
  properties: DisplaySettings["properties"];
  fixedTeamId: string | undefined;
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  fetchNextPage: () => void;
};

function VirtualIssueList({
  groups,
  issues,
  showHeaders,
  nestSubIssues: shouldNest,
  properties,
  fixedTeamId,
  hasNextPage,
  isFetchingNextPage,
  fetchNextPage,
}: TVirtualProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scrollElement, setScrollElement] = useState<HTMLElement | null>(null);
  const [scrollMargin, setScrollMargin] = useState(0);
  const [collapsed, setCollapsed] = useState<string[]>([]);
  const { canEdit } = useWorkspace();
  const openNewIssue = useNewIssueStore((state) => state.open);
  const peek = useIssuePeek();
  const isPeekable = useIsPeekable();
  const openPage = useOpenIssuePage();
  const selectedIds = useIssueSelectionStore((state) => state.selectedIds);
  const toggle = useIssueSelectionStore((state) => state.toggle);
  const selectRange = useIssueSelectionStore((state) => state.selectRange);
  const setFocused = useIssueSelectionStore((state) => state.setFocused);
  const clear = useIssueSelectionStore((state) => state.clear);

  const items = useMemo<TListItem[]>(
    () =>
      groups.flatMap((group) => {
        const header: TListItem[] = showHeaders
          ? [{ kind: "header", key: `group:${group.key}`, group }]
          : [];
        if (showHeaders && collapsed.includes(group.key)) return header;
        const rows = shouldNest
          ? nestSubIssues(group.issues)
          : group.issues.map((issue) => ({ issue, depth: 0 }));
        return [
          ...header,
          ...rows.map(({ issue, depth }): TListItem => ({
            kind: "row",
            key: `${group.key}:${issue.id}`,
            issue,
            depth,
          })),
        ];
      }),
    [groups, showHeaders, shouldNest, collapsed],
  );
  const orderedIds = useMemo(
    () => items.flatMap((item) => (item.kind === "row" ? [item.issue.id] : [])),
    [items],
  );

  useLayoutEffect(() => {
    const container = containerRef.current;
    const parent = findScrollParent(container);
    setScrollElement(parent);
    if (!container || !parent) return;
    const measure = () =>
      setScrollMargin(
        container.getBoundingClientRect().top -
          parent.getBoundingClientRect().top +
          parent.scrollTop,
      );
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(parent);
    return () => observer.disconnect();
  }, []);

  useEffect(() => () => clear(), [clear]);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => scrollElement,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
    scrollMargin,
    getItemKey: (index) => items[index].key,
  });
  const virtualItems = virtualizer.getVirtualItems();
  const lastIndex = virtualItems[virtualItems.length - 1]?.index ?? 0;

  useEffect(() => {
    if (!hasNextPage || isFetchingNextPage) return;
    if (lastIndex < items.length - PREFETCH_ROWS) return;
    fetchNextPage();
  }, [lastIndex, items.length, hasNextPage, isFetchingNextPage, fetchNextPage]);

  useHotkeyScope([
    {
      hotkey: "X",
      callback: () => {
        const focusedId = useIssueSelectionStore.getState().focusedId;
        if (focusedId && orderedIds.includes(focusedId)) toggle(focusedId);
      },
    },
    { hotkey: "Escape", callback: clear, enabled: selectedIds.length > 0, ignoreInputs: true },
  ]);

  const handleClick = (event: MouseEvent, issue: TIssue) => {
    if (event.metaKey || event.ctrlKey) {
      void openPage(issue.identifier);
      return;
    }
    if (event.shiftKey) {
      selectRange(orderedIds, issue.id);
      return;
    }
    if (peek && isPeekable) {
      peek.openPeek(issue.identifier);
      return;
    }
    void openPage(issue.identifier);
  };

  const selectedIssues = issues.filter((issue) => selectedIds.includes(issue.id));
  const baseDefaults = fixedTeamId ? { teamId: fixedTeamId } : {};

  return (
    <>
      <div ref={containerRef}>
        <List>
          <div className="relative w-full" style={{ height: virtualizer.getTotalSize() }}>
            {virtualItems.map((virtualItem) => {
              const item = items[virtualItem.index];
              return (
                <div
                  key={virtualItem.key}
                  className="absolute inset-x-0 top-0"
                  style={{ transform: `translateY(${virtualItem.start - scrollMargin}px)` }}
                >
                  {item.kind === "header" ? (
                    <IssueGroupHeader
                      group={item.group}
                      isCollapsed={collapsed.includes(item.group.key)}
                      onToggle={() =>
                        setCollapsed((current) =>
                          current.includes(item.group.key)
                            ? current.filter((key) => key !== item.group.key)
                            : [...current, item.group.key],
                        )
                      }
                      onCreate={
                        canEdit
                          ? () => openNewIssue(newIssueDefaultsForGroup(item.group, baseDefaults))
                          : undefined
                      }
                    />
                  ) : (
                    <IssueRow
                      issue={item.issue}
                      depth={item.depth}
                      properties={properties}
                      isSelected={selectedIds.includes(item.issue.id)}
                      isSelecting={selectedIds.length > 0}
                      isActive={peek?.peek?.toUpperCase() === item.issue.identifier}
                      canEdit={canEdit}
                      onClick={handleClick}
                      onToggleSelect={(issue, extendRange) =>
                        extendRange ? selectRange(orderedIds, issue.id) : toggle(issue.id)
                      }
                      onFocus={(issue) => setFocused(issue.id)}
                    />
                  )}
                </div>
              );
            })}
          </div>
        </List>
      </div>
      <BulkActionBar issues={selectedIssues} onClear={clear} />
    </>
  );
}
