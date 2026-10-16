import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
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
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { BoxIcon, GripVerticalIcon, PencilIcon, PlusIcon } from "lucide-react";
import { useState } from "react";

import { sortOrderBetween } from "@/components/issues/sort-order";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { List } from "@/components/layout/list";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { DeleteProjectStatusDialog } from "@/components/projects/delete-project-status-dialog";
import { ProjectStatusDialog } from "@/components/projects/project-status-dialog";
import {
  PROJECT_STATUS_TYPES,
  ProjectStatusIcon,
  projectStatusTypeLabels,
} from "@/components/projects/project-status-icon";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  projectStatusesQuery,
  type TProjectStatus,
  type TProjectStatusType,
  useUpdateProjectStatusMutation,
} from "@/lib/queries/projects";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/project-statuses")({
  component: ProjectStatusesSettingsPage,
});

const typeDescriptions: Record<TProjectStatusType, string> = {
  backlog: "Ideas nobody has planned yet.",
  planned: "Projects scheduled to start.",
  started: "Projects in progress.",
  paused: "Projects on hold.",
  completed: "Finished projects.",
  canceled: "Projects that will not be done.",
};

function byPosition(a: TProjectStatus, b: TProjectStatus) {
  return a.position - b.position;
}

function ProjectStatusesSettingsPage() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(projectStatusesQuery(workspace.id));
  const statuses = data?.statuses ?? [];

  return (
    <SettingsPage
      title="Project statuses"
      description="The statuses a project moves through. Drag to reorder within a group."
    >
      {isPending && <ListSkeleton />}
      {!isPending && error && <ErrorCard message={error.message} onRetry={() => void refetch()} />}
      {!isPending && !error && statuses.length === 0 && (
        <EmptyState icon={BoxIcon} title="No project statuses yet" />
      )}
      {!isPending &&
        !error &&
        statuses.length > 0 &&
        PROJECT_STATUS_TYPES.map((type) => (
          <StatusGroup
            key={type}
            workspaceId={workspace.id}
            type={type}
            statuses={statuses.filter((status) => status.type === type).sort(byPosition)}
            allStatuses={statuses}
          />
        ))}
    </SettingsPage>
  );
}

function StatusGroup({
  workspaceId,
  type,
  statuses,
  allStatuses,
}: {
  workspaceId: string;
  type: TProjectStatusType;
  statuses: TProjectStatus[];
  allStatuses: TProjectStatus[];
}) {
  const [dragOrder, setDragOrder] = useState<string[] | null>(null);
  const { mutate: updateStatus } = useUpdateProjectStatusMutation();
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );
  const ordered = dragOrder
    ? dragOrder
        .map((id) => statuses.find((status) => status.id === id))
        .filter((status): status is TProjectStatus => !!status)
    : statuses;
  const lastPosition = statuses.at(-1)?.position ?? 0;

  const onDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const ids = ordered.map((status) => status.id);
    const next = arrayMove(ids, ids.indexOf(String(active.id)), ids.indexOf(String(over.id)));
    const index = next.indexOf(String(active.id));
    const before = ordered.find((status) => status.id === next[index - 1]);
    const after = ordered.find((status) => status.id === next[index + 1]);
    setDragOrder(next);
    updateStatus(
      {
        status_id: String(active.id),
        position: sortOrderBetween(before?.position, after?.position),
      },
      { onSettled: () => setDragOrder(null) },
    );
  };

  return (
    <SettingsSection
      title={projectStatusTypeLabels[type]}
      description={typeDescriptions[type]}
      actions={
        <ProjectStatusDialog
          mode="create"
          workspaceId={workspaceId}
          type={type}
          position={lastPosition + 1}
          trigger={
            <Button
              variant="ghost"
              size="icon"
              aria-label={`Add ${projectStatusTypeLabels[type].toLowerCase()} status`}
            >
              <PlusIcon className="size-4" />
            </Button>
          }
        />
      }
    >
      {ordered.length === 0 ? (
        <p className="text-muted-foreground text-sm">No statuses of this type.</p>
      ) : (
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
          <SortableContext
            items={ordered.map((status) => status.id)}
            strategy={verticalListSortingStrategy}
          >
            <List>
              {ordered.map((status) => (
                <SortableStatusRow key={status.id} status={status} allStatuses={allStatuses} />
              ))}
            </List>
          </SortableContext>
        </DndContext>
      )}
    </SettingsSection>
  );
}

function SortableStatusRow({
  status,
  allStatuses,
}: {
  status: TProjectStatus;
  allStatuses: TProjectStatus[];
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: status.id });
  return (
    <div
      ref={setNodeRef}
      role="listitem"
      data-dragging={isDragging || undefined}
      style={{
        transform: transform ? `translate3d(${transform.x}px, ${transform.y}px, 0)` : undefined,
        transition,
      }}
      className="bg-card data-dragging:bg-accent relative flex h-9 w-full items-center gap-2 border-b px-2 text-sm last:border-b-0 data-dragging:z-10"
    >
      <Button
        ref={setActivatorNodeRef}
        variant="ghost"
        size="icon"
        aria-label={`Reorder ${status.name}`}
        className="text-muted-foreground cursor-grab"
        {...attributes}
        {...listeners}
      >
        <GripVerticalIcon className="size-4" />
      </Button>
      <ProjectStatusIcon type={status.type} color={status.color} />
      <span className="truncate">{status.name}</span>
      {status.description && (
        <span className="text-muted-foreground hidden truncate text-xs sm:inline">
          {status.description}
        </span>
      )}
      <div className="ml-auto flex shrink-0 items-center gap-2">
        <ProjectStatusDialog
          mode="edit"
          status={status}
          trigger={
            <Button variant="ghost" size="icon" aria-label={`Edit ${status.name}`}>
              <PencilIcon className="size-4" />
            </Button>
          }
        />
        <DeleteProjectStatusDialog status={status} statuses={allStatuses} />
      </div>
    </div>
  );
}
