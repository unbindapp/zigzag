import { closestCenter, DndContext, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useQuery } from "@tanstack/react-query";
import { ChevronRightIcon, FolderIcon, GripVerticalIcon, PlusIcon, TagIcon } from "lucide-react";
import { type ReactNode, useState } from "react";

import { LabelDialog } from "@/components/labels/label-dialog";
import { LabelRowMenu } from "@/components/labels/label-row-menu";
import { reorderedPosition, sortableStyle, useSortSensors } from "@/components/labels/sortable";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsSection } from "@/components/layout/settings-section";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Switch } from "@/components/ui/switch";
import {
  labelTreeQuery,
  type TLabel,
  type TLabelTree,
  useUpdateLabelMutation,
} from "@/lib/queries/labels";

type TProps = {
  workspaceId: string;
  teamId?: string;
  canEdit: boolean;
  title: string;
  description: string;
};

export function LabelManager({ workspaceId, teamId, canEdit, title, description }: TProps) {
  const [showArchived, setShowArchived] = useState(false);
  const [creating, setCreating] = useState<"label" | "group" | null>(null);
  const { data, isPending, error, refetch } = useQuery(
    labelTreeQuery({ workspaceId, teamId, includeArchived: showArchived }),
  );
  const labels = (data?.labels ?? []).filter(
    (label) => (label.team_id ?? null) === (teamId ?? null),
  );
  const groups = labels.filter((label) => label.is_group && !label.archived_at);

  const createMenu = canEdit && (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button size="sm">
            <PlusIcon className="size-4" />
            New
          </Button>
        }
      />
      <DropdownMenuContent align="end" className="w-40">
        <DropdownMenuItem onClick={() => setCreating("label")}>
          <TagIcon className="size-4" />
          New label
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setCreating("group")}>
          <FolderIcon className="size-4" />
          New group
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );

  return (
    <SettingsSection
      title={title}
      description={description}
      actions={
        <>
          <ArchivedToggle checked={showArchived} onChange={setShowArchived} />
          {createMenu}
        </>
      }
    >
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={() => void refetch()}
        isEmpty={labels.length === 0}
        empty={{
          icon: TagIcon,
          title: "No labels yet",
          description: canEdit
            ? "Labels sort issues by type, area or anything else. Groups hold related labels."
            : "An admin can add labels here.",
        }}
      >
        <SortableLabels items={labels} groups={groups} canEdit={canEdit} />
      </ListStates>
      {canEdit && (
        <LabelDialog
          mode="create"
          workspaceId={workspaceId}
          teamId={teamId}
          isGroup={creating === "group"}
          groups={groups}
          open={creating !== null}
          onOpenChange={(open) => {
            if (!open) setCreating(null);
          }}
        />
      )}
    </SettingsSection>
  );
}

function ArchivedToggle({
  checked,
  onChange,
}: {
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label className="text-muted-foreground flex items-center gap-2 text-xs">
      <Switch checked={checked} onCheckedChange={onChange} aria-label="Show archived labels" />
      Archived
    </label>
  );
}

function SortableLabels({
  items,
  groups,
  canEdit,
}: {
  items: TLabelTree[];
  groups: TLabelTree[];
  canEdit: boolean;
}) {
  return (
    <SortableList items={items}>
      {(ordered) => (
        <List>
          {ordered.map((label, index) =>
            label.is_group ? (
              <GroupBlock
                key={label.id}
                group={label}
                groups={groups}
                canEdit={canEdit}
                isLast={index === ordered.length - 1}
              />
            ) : (
              <SortableLabelRow
                key={label.id}
                label={label}
                groups={groups}
                canEdit={canEdit}
                divided={index < ordered.length - 1}
              />
            ),
          )}
        </List>
      )}
    </SortableList>
  );
}

function SortableList<T extends TLabel>({
  items,
  children,
}: {
  items: T[];
  children: (ordered: T[]) => ReactNode;
}) {
  const [dragOrder, setDragOrder] = useState<string[] | null>(null);
  const { mutate: updateLabel } = useUpdateLabelMutation();
  const sensors = useSortSensors();
  const ordered = dragOrder
    ? dragOrder
        .map((id) => items.find((item) => item.id === id))
        .filter((item): item is T => !!item)
    : items;

  const onDragEnd = (event: DragEndEvent) => {
    const moved = reorderedPosition(ordered, event);
    if (!moved) return;
    setDragOrder(moved.order);
    updateLabel(
      { label_id: moved.id, sort_order: moved.sortOrder },
      { onSettled: () => setDragOrder(null) },
    );
  };

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}>
      <SortableContext
        items={ordered.map((item) => item.id)}
        strategy={verticalListSortingStrategy}
      >
        {children(ordered)}
      </SortableContext>
    </DndContext>
  );
}

function GroupBlock({
  group,
  groups,
  canEdit,
  isLast,
}: {
  group: TLabelTree;
  groups: TLabelTree[];
  canEdit: boolean;
  isLast: boolean;
}) {
  const [isOpen, setIsOpen] = useState(true);
  const isExpanded = isOpen && group.children.length > 0;
  const sortable = useSortable({ id: group.id, disabled: !canEdit });
  return (
    <div
      ref={sortable.setNodeRef}
      role="listitem"
      data-dragging={sortable.isDragging || undefined}
      style={sortableStyle(sortable.transform, sortable.transition)}
      className="bg-card data-dragging:bg-accent relative flex w-full flex-col data-dragging:z-10"
    >
      <LabelRowBody
        label={group}
        divided={isExpanded || !isLast}
        groups={groups}
        canEdit={canEdit}
        handle={sortable}
        leading={
          <Button
            variant="ghost"
            size="icon"
            aria-expanded={isOpen}
            aria-label={isOpen ? `Collapse ${group.name}` : `Expand ${group.name}`}
            onClick={() => setIsOpen((open) => !open)}
          >
            <ChevronRightIcon
              data-open={isOpen || undefined}
              className="size-4 transition-transform data-open:rotate-90"
            />
          </Button>
        }
        meta={
          <>
            <span className="text-muted-foreground text-xs">
              {group.children.length} {group.children.length === 1 ? "label" : "labels"}
            </span>
            {group.group_single_select && <Badge variant="muted">Single select</Badge>}
          </>
        }
      />
      {isExpanded && (
        <SortableList items={group.children}>
          {(ordered) => (
            <div role="list" className="flex flex-col">
              {ordered.map((child, index) => (
                <SortableLabelRow
                  key={child.id}
                  label={child}
                  groups={groups}
                  canEdit={canEdit}
                  divided={!isLast || index < ordered.length - 1}
                  nested
                />
              ))}
            </div>
          )}
        </SortableList>
      )}
    </div>
  );
}

function SortableLabelRow({
  label,
  groups,
  canEdit,
  divided,
  nested,
}: {
  label: TLabel;
  groups: TLabelTree[];
  canEdit: boolean;
  divided: boolean;
  nested?: boolean;
}) {
  const sortable = useSortable({ id: label.id, disabled: !canEdit });
  return (
    <div
      ref={sortable.setNodeRef}
      role="listitem"
      data-dragging={sortable.isDragging || undefined}
      style={sortableStyle(sortable.transform, sortable.transition)}
      className="bg-card data-dragging:bg-accent relative flex w-full data-dragging:z-10"
    >
      <LabelRowBody
        label={label}
        divided={divided}
        groups={groups}
        canEdit={canEdit}
        handle={sortable}
        leading={nested ? <span aria-hidden className="size-8 shrink-0" /> : undefined}
      />
    </div>
  );
}

function LabelRowBody({
  label,
  groups,
  canEdit,
  handle,
  leading,
  meta,
  divided,
}: {
  label: TLabel;
  groups: TLabelTree[];
  canEdit: boolean;
  handle: ReturnType<typeof useSortable>;
  leading?: ReactNode;
  meta?: ReactNode;
  divided: boolean;
}) {
  return (
    <div
      data-archived={label.archived_at ? true : undefined}
      data-divided={divided || undefined}
      className="data-archived:text-muted-foreground flex h-9 w-full items-center gap-2 px-2 text-sm data-divided:border-b"
    >
      {canEdit && (
        <Button
          ref={handle.setActivatorNodeRef}
          variant="ghost"
          size="icon"
          aria-label={`Reorder ${label.name}`}
          className="text-muted-foreground cursor-grab"
          {...handle.attributes}
          {...handle.listeners}
        >
          <GripVerticalIcon className="size-4" />
        </Button>
      )}
      {leading}
      <span
        aria-hidden
        className="size-2.5 shrink-0 rounded-full"
        style={{ backgroundColor: label.color }}
      />
      <span className="truncate">{label.name}</span>
      {label.description && (
        <span className="text-muted-foreground hidden truncate text-xs sm:inline">
          {label.description}
        </span>
      )}
      {meta}
      {label.archived_at && <Badge variant="muted">Archived</Badge>}
      {canEdit && (
        <div className="ml-auto flex shrink-0 items-center gap-2">
          <LabelRowMenu label={label} groups={groups} />
        </div>
      )}
    </div>
  );
}
