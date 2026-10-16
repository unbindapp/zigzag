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
import { Link } from "@tanstack/react-router";
import { BoxIcon, PlusIcon } from "lucide-react";
import { type CSSProperties, type HTMLAttributes, type ReactNode, type Ref, useState } from "react";

import { sortOrderAtIndex } from "@/components/issues/sort-order";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { HealthBadge } from "@/components/projects/health-badge";
import { useNewProjectStore } from "@/components/projects/new-project-store";
import { formatProjectDate } from "@/components/projects/project-dates";
import { type TProjectGroup } from "@/components/projects/project-groups";
import { type TProjectListFilters, useProjectList } from "@/components/projects/project-list";
import { ProgressBar } from "@/components/projects/progress-bar";
import { ProjectStatusIcon } from "@/components/projects/project-status-icon";
import { Avatar } from "@/components/shared/avatar";
import { PriorityIcon } from "@/components/shared/priority-icon";
import { ProjectIcon } from "@/components/shared/project-icon";
import { Button } from "@/components/ui/button";
import { cn } from "@/components/ui/utils";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TProjectListItem, useUpdateProjectMutation } from "@/lib/queries/projects";

const DRAG_ACTIVATION_DISTANCE = 6;

type TCardData = { statusId: string; projectId: string };

type TProps = {
  workspaceId: string;
  filters: TProjectListFilters;
  emptyState?: ReactNode;
};

export function ProjectBoard({ workspaceId, filters, emptyState }: TProps) {
  const { groups, projects, isPending, error, refetch } = useProjectList(workspaceId, {
    ...filters,
    order_by: "sort_order",
  });
  if (isPending) return <ListSkeleton rows={8} />;
  if (error) return <ErrorCard message={error.message} onRetry={refetch} />;
  if (projects.length === 0) {
    return (
      emptyState ?? (
        <EmptyState
          icon={BoxIcon}
          title="No projects yet"
          description="Projects group issues toward a goal."
        />
      )
    );
  }
  return <Board groups={groups} filters={filters} />;
}

function Board({ groups, filters }: { groups: TProjectGroup[]; filters: TProjectListFilters }) {
  const { canEdit } = useWorkspace();
  const { mutate: updateProject } = useUpdateProjectMutation();
  const [active, setActive] = useState<TCardData | null>(null);
  const [pending, setPending] = useState<{ projectId: string; statusId: string } | null>(null);
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: DRAG_ACTIVATION_DISTANCE } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );
  const shown = pending ? moveProject(groups, pending.projectId, pending.statusId) : groups;
  const activeProject = active
    ? shown.flatMap((group) => group.projects).find((project) => project.id === active.projectId)
    : undefined;

  const onDragStart = (event: DragStartEvent) =>
    setActive((event.active.data.current as TCardData | undefined) ?? null);

  const onDragEnd = (event: DragEndEvent) => {
    setActive(null);
    const source = event.active.data.current as TCardData | undefined;
    if (!source || !event.over) return;
    const overCard = event.over.data.current as Partial<TCardData> | undefined;
    const targetId = overCard?.statusId ?? String(event.over.id).replace(/^column::/, "");
    const sourceGroup = shown.find((group) => group.status.id === source.statusId);
    const targetGroup = shown.find((group) => group.status.id === targetId);
    const project = sourceGroup?.projects.find((item) => item.id === source.projectId);
    if (!sourceGroup || !targetGroup || !project) return;

    const isSameColumn = sourceGroup.status.id === targetGroup.status.id;
    const overIndex = overCard?.projectId
      ? targetGroup.projects.findIndex((item) => item.id === overCard.projectId)
      : targetGroup.projects.length;
    const reordered = isSameColumn
      ? arrayMove(
          targetGroup.projects,
          targetGroup.projects.findIndex((item) => item.id === project.id),
          overIndex,
        )
      : [
          ...targetGroup.projects.slice(0, overIndex),
          project,
          ...targetGroup.projects.slice(overIndex),
        ];
    const newIndex = reordered.findIndex((item) => item.id === project.id);
    if (
      isSameColumn &&
      newIndex === sourceGroup.projects.findIndex((item) => item.id === project.id)
    )
      return;
    const neighbors = reordered
      .filter((item) => item.id !== project.id)
      .map((item) => item.sort_order);

    setPending({ projectId: project.id, statusId: targetGroup.status.id });
    updateProject(
      {
        project_id: project.id,
        sort_order: sortOrderAtIndex(neighbors, newIndex),
        ...(isSameColumn ? {} : { status_id: targetGroup.status.id }),
      },
      { onSettled: () => setPending(null) },
    );
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={onDragStart}
      onDragCancel={() => setActive(null)}
      onDragEnd={onDragEnd}
    >
      <div className="flex w-full min-w-0 items-start gap-3 overflow-x-auto pb-2">
        {shown.map((group) => (
          <BoardColumn key={group.status.id} group={group} canEdit={canEdit} filters={filters} />
        ))}
      </div>
      <DragOverlay>
        {activeProject && <ProjectCard project={activeProject} isOverlay />}
      </DragOverlay>
    </DndContext>
  );
}

function moveProject(groups: TProjectGroup[], projectId: string, statusId: string) {
  const project = groups.flatMap((group) => group.projects).find((item) => item.id === projectId);
  if (!project) return groups;
  return groups.map((group) => {
    const rest = group.projects.filter((item) => item.id !== projectId);
    return { ...group, projects: group.status.id === statusId ? [...rest, project] : rest };
  });
}

function BoardColumn({
  group,
  canEdit,
  filters,
}: {
  group: TProjectGroup;
  canEdit: boolean;
  filters: TProjectListFilters;
}) {
  const { setNodeRef, isOver } = useDroppable({ id: `column::${group.status.id}` });
  const openNewProject = useNewProjectStore((state) => state.open);
  return (
    <section
      aria-label={group.status.name}
      data-over={isOver || undefined}
      className="bg-muted data-over:border-primary/5-10 flex w-72 shrink-0 flex-col rounded-lg border"
    >
      <header className="flex h-10 items-center gap-2 px-3 text-sm">
        <ProjectStatusIcon type={group.status.type} color={group.status.color} />
        <span className="min-w-0 truncate font-medium">{group.status.name}</span>
        <span className="text-muted-foreground text-xs">{group.projects.length}</span>
        <span className="flex-1" />
        {canEdit && (
          <Button
            variant="ghost"
            size="icon"
            aria-label={`New project in ${group.status.name}`}
            onClick={() =>
              openNewProject({
                statusId: group.status.id,
                teamId: filters.team_id,
                initiativeId: filters.initiative_id,
              })
            }
          >
            <PlusIcon className="size-4" />
          </Button>
        )}
      </header>
      <SortableContext
        id={group.status.id}
        items={group.projects.map((project) => project.id)}
        strategy={verticalListSortingStrategy}
      >
        <div ref={setNodeRef} role="list" className="flex min-h-16 flex-col gap-3 px-2 pb-2">
          {group.projects.map((project) => (
            <SortableProjectCard
              key={project.id}
              statusId={group.status.id}
              project={project}
              canEdit={canEdit}
            />
          ))}
        </div>
      </SortableContext>
    </section>
  );
}

function SortableProjectCard({
  statusId,
  project,
  canEdit,
}: {
  statusId: string;
  project: TProjectListItem;
  canEdit: boolean;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: project.id,
    data: { statusId, projectId: project.id } satisfies TCardData,
    disabled: !canEdit,
  });
  return (
    <ProjectCard
      ref={setNodeRef}
      project={project}
      isDragging={isDragging}
      style={{
        transform: transform ? `translate3d(${transform.x}px, ${transform.y}px, 0)` : undefined,
        transition,
      }}
      dragProps={{ ...attributes, ...listeners }}
    />
  );
}

function ProjectCard({
  project,
  isDragging,
  isOverlay,
  ref,
  style,
  dragProps,
}: {
  project: TProjectListItem;
  isDragging?: boolean;
  isOverlay?: boolean;
  ref?: Ref<HTMLDivElement>;
  style?: CSSProperties;
  dragProps?: HTMLAttributes<HTMLDivElement>;
}) {
  const { workspace } = useWorkspace();
  const lead = project.lead;
  return (
    <div
      ref={ref}
      style={style}
      {...dragProps}
      role="listitem"
      aria-label={project.name}
      data-dragging={isDragging || undefined}
      className={cn(
        "bg-card has-hover:hover:border-primary/4-10 focus-visible:ring-primary/8-10 flex w-full cursor-default flex-col gap-2 rounded-lg border p-4 text-sm outline-hidden focus-visible:ring-1 data-dragging:opacity-40",
        isOverlay && "shadow-lg",
      )}
    >
      <div className="flex min-w-0 items-center gap-1.5">
        <ProjectIcon icon={project.icon} color={project.color} />
        <Link
          to="/$workspace/project/$slug"
          params={{ workspace: workspace.slug, slug: project.slug }}
          className="min-w-0 flex-1 truncate font-medium"
        >
          {project.name}
        </Link>
        <PriorityIcon priority={project.priority} />
      </div>
      {project.summary && (
        <p className="text-muted-foreground line-clamp-2 text-xs">{project.summary}</p>
      )}
      <ProgressBar percent={project.progress.percent} label={`${project.name} progress`} />
      <div className="flex items-center gap-2">
        {project.health && <HealthBadge health={project.health} />}
        <span className="text-muted-foreground min-w-0 flex-1 truncate text-xs">
          {project.target_date
            ? formatProjectDate(project.target_date, project.target_date_resolution)
            : ""}
        </span>
        {lead && (
          <Avatar
            id={lead.id}
            name={lead.name || lead.display_name}
            src={lead.avatar_url}
            size="sm"
          />
        )}
      </div>
    </div>
  );
}
