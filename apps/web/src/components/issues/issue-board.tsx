import {
  closestCorners,
  DndContext,
  type DragEndEvent,
  DragOverlay,
  type DragStartEvent,
  KeyboardSensor,
  PointerSensor,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { LayoutGridIcon, PlusIcon } from "lucide-react";
import { type MouseEvent, type ReactNode, useState } from "react";

import type { DisplaySettings } from "@/components/issues/display";
import type { ListIssuesFilters } from "@/components/issues/filters";
import { GroupIcon } from "@/components/issues/group-icon";
import type { TIssueGroup } from "@/components/issues/grouping";
import { patchToOptimistic, type TIssuePatch } from "@/components/issues/issue-draft";
import { IssueCard } from "@/components/issues/issue-card";
import { useIsPeekable, useIssuePeek } from "@/components/issues/issue-peek";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { sortOrderAtIndex } from "@/components/issues/sort-order";
import { issuePatchToUpdateInput, type TIssue, useUpdateIssueMutation } from "@/lib/queries/issues";
import {
  newIssueDefaultsForGroup,
  useIssueCollection,
} from "@/components/issues/use-issue-collection";
import { useOpenIssuePage } from "@/components/issues/use-issue-actions";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export type TIssueBoardProps = {
  workspaceId: string;
  filters: ListIssuesFilters;
  display?: Partial<DisplaySettings>;
  fixedFilters?: Partial<ListIssuesFilters>;
  emptyState?: ReactNode;
};

const DRAG_ACTIVATION_DISTANCE = 6;

type TCardData = { groupKey: string; issueId: string };

function cardId(groupKey: string, issueId: string) {
  return `${groupKey}::${issueId}`;
}

function columnId(groupKey: string) {
  return `column::${groupKey}`;
}

function groupPatch(group: TIssueGroup, issue: TIssue): TIssuePatch | "unsupported" {
  const meta = group.meta;
  switch (meta.kind) {
    case "state":
      if (meta.state.team_id !== issue.team_id) return "unsupported";
      return { state: meta.state };
    case "priority":
      return { priority: meta.priority };
    case "assignee":
      return { assignee: meta.user };
    case "project":
      return { project: meta.project };
    case "cycle":
      if (issue.cycle?.id === meta.cycle?.id) return {};
      return { cycle: meta.cycle };
    case "none":
    case "label":
    case "parent":
      return "unsupported";
  }
}

export function IssueBoard(props: TIssueBoardProps) {
  const { query, issues, groups, display, fixedTeamId } = useIssueCollection({
    ...props,
    display: { group_by: "state", ...props.display },
  });

  if (query.isPending) return <ListSkeleton rows={8} />;
  if (query.error) {
    return <ErrorCard message={query.error.message} onRetry={() => void query.refetch()} />;
  }
  if (issues.length === 0 && groups.length === 0) {
    return (
      props.emptyState ?? (
        <EmptyState
          icon={LayoutGridIcon}
          title="No issues yet"
          description="Issues that match these filters show up here."
        />
      )
    );
  }
  return (
    <Board
      groups={groups}
      properties={display.properties}
      isManualOrder={display.order_by === "manual"}
      fixedTeamId={fixedTeamId}
    />
  );
}

function Board({
  groups,
  properties,
  isManualOrder,
  fixedTeamId,
}: {
  groups: TIssueGroup[];
  properties: DisplaySettings["properties"];
  isManualOrder: boolean;
  fixedTeamId: string | undefined;
}) {
  const { canEdit } = useWorkspace();
  const { mutate: updateIssue } = useUpdateIssueMutation();
  const [activeCard, setActiveCard] = useState<TCardData | null>(null);
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: DRAG_ACTIVATION_DISTANCE } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );
  const activeIssue = activeCard
    ? groups
        .find((group) => group.key === activeCard.groupKey)
        ?.issues.find((issue) => issue.id === activeCard.issueId)
    : undefined;

  const handleDragStart = (event: DragStartEvent) => {
    setActiveCard((event.active.data.current as TCardData | undefined) ?? null);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveCard(null);
    const source = event.active.data.current as TCardData | undefined;
    if (!source || !event.over) return;
    const overCard = event.over.data.current as Partial<TCardData> | undefined;
    const targetKey = overCard?.groupKey ?? String(event.over.id).replace(/^column::/, "");
    const sourceGroup = groups.find((group) => group.key === source.groupKey);
    const targetGroup = groups.find((group) => group.key === targetKey);
    const issue = sourceGroup?.issues.find((item) => item.id === source.issueId);
    if (!sourceGroup || !targetGroup || !issue) return;

    const isSameColumn = sourceGroup.key === targetGroup.key;
    const overIndex = overCard?.issueId
      ? targetGroup.issues.findIndex((item) => item.id === overCard.issueId)
      : targetGroup.issues.length;
    const reordered = isSameColumn
      ? arrayMove(
          targetGroup.issues,
          targetGroup.issues.findIndex((item) => item.id === issue.id),
          overIndex,
        )
      : [...targetGroup.issues.slice(0, overIndex), issue, ...targetGroup.issues.slice(overIndex)];
    const newIndex = reordered.findIndex((item) => item.id === issue.id);
    const neighbors = reordered
      .filter((item) => item.id !== issue.id)
      .map((item) => item.sort_order);

    const patch = isSameColumn ? {} : groupPatch(targetGroup, issue);
    if (patch === "unsupported") {
      toast.add({
        type: "warning",
        title: "Cannot move there",
        description: "Change this property from the issue instead.",
      });
      return;
    }
    if (isSameColumn && newIndex === sourceGroup.issues.findIndex((item) => item.id === issue.id)) {
      return;
    }
    const sortOrder =
      isManualOrder || isSameColumn ? sortOrderAtIndex(neighbors, newIndex) : undefined;
    updateIssue({
      input: {
        ...issuePatchToUpdateInput(issue.id, patch),
        ...(sortOrder !== undefined ? { sort_order: sortOrder } : {}),
      },
      optimistic: {
        ...patchToOptimistic(patch),
        ...(sortOrder !== undefined ? { sort_order: sortOrder } : {}),
      },
    });
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragCancel={() => setActiveCard(null)}
      onDragEnd={handleDragEnd}
    >
      <div className="flex w-full min-w-0 items-start gap-3 overflow-x-auto pb-2">
        {groups.map((group) => (
          <BoardColumn
            key={group.key}
            group={group}
            properties={properties}
            canEdit={canEdit}
            fixedTeamId={fixedTeamId}
          />
        ))}
      </div>
      <DragOverlay>
        {activeIssue && (
          <IssueCard issue={activeIssue} properties={properties} canEdit={false} isOverlay />
        )}
      </DragOverlay>
    </DndContext>
  );
}

function BoardColumn({
  group,
  properties,
  canEdit,
  fixedTeamId,
}: {
  group: TIssueGroup;
  properties: DisplaySettings["properties"];
  canEdit: boolean;
  fixedTeamId: string | undefined;
}) {
  const { setNodeRef, isOver } = useDroppable({ id: columnId(group.key) });
  const openNewIssue = useNewIssueStore((state) => state.open);
  const ids = group.issues.map((issue) => cardId(group.key, issue.id));
  return (
    <section
      aria-label={group.label}
      data-over={isOver || undefined}
      className="bg-muted data-over:border-primary/5-10 flex w-72 shrink-0 flex-col rounded-lg border"
    >
      <header className="flex h-10 items-center gap-2 px-3 text-sm">
        <GroupIcon meta={group.meta} />
        <span className="min-w-0 truncate font-medium">{group.label}</span>
        <span className="text-muted-foreground text-xs">{group.issues.length}</span>
        <span className="flex-1" />
        {canEdit && (
          <Button
            variant="ghost"
            size="icon"
            aria-label={`New issue in ${group.label}`}
            onClick={() =>
              openNewIssue(
                newIssueDefaultsForGroup(group, fixedTeamId ? { teamId: fixedTeamId } : {}),
              )
            }
          >
            <PlusIcon className="size-4" />
          </Button>
        )}
      </header>
      <SortableContext id={group.key} items={ids} strategy={verticalListSortingStrategy}>
        <div ref={setNodeRef} role="list" className="flex min-h-16 flex-col gap-3 px-2 pb-2">
          {group.issues.map((issue) => (
            <SortableIssueCard
              key={issue.id}
              groupKey={group.key}
              issue={issue}
              properties={properties}
              canEdit={canEdit}
            />
          ))}
        </div>
      </SortableContext>
    </section>
  );
}

function SortableIssueCard({
  groupKey,
  issue,
  properties,
  canEdit,
}: {
  groupKey: string;
  issue: TIssue;
  properties: DisplaySettings["properties"];
  canEdit: boolean;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: cardId(groupKey, issue.id),
    data: { groupKey, issueId: issue.id } satisfies TCardData,
    disabled: !canEdit,
  });
  const peek = useIssuePeek();
  const isPeekable = useIsPeekable();
  const openPage = useOpenIssuePage();
  const handleClick = (event: MouseEvent) => {
    if (event.metaKey || event.ctrlKey || !peek || !isPeekable) {
      void openPage(issue.identifier);
      return;
    }
    peek.openPeek(issue.identifier);
  };
  return (
    <IssueCard
      ref={setNodeRef}
      issue={issue}
      properties={properties}
      canEdit={canEdit}
      isDragging={isDragging}
      isActive={peek?.peek?.toUpperCase() === issue.identifier}
      onClick={handleClick}
      style={{
        transform: transform ? `translate3d(${transform.x}px, ${transform.y}px, 0)` : undefined,
        transition,
      }}
      dragProps={{ ...attributes, ...listeners }}
    />
  );
}
