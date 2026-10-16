import { closestCenter, DndContext, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useQuery } from "@tanstack/react-query";
import { type MouseEvent, useRef, useState } from "react";

import { favoriteKindIcons, favoriteLink } from "@/components/favorites/favorite-link";
import { reorderedPosition, sortableStyle, useSortSensors } from "@/components/labels/sortable";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { SidebarNavItem } from "@/components/workspace/sidebar-nav";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  favoritesListQuery,
  type TFavorite,
  useReorderFavoritesMutation,
} from "@/lib/queries/favorites";

type TProps = {
  onNavigate?: () => void;
};

export function SidebarFavorites({ onNavigate }: TProps) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(favoritesListQuery(workspace.id));
  const [dragOrder, setDragOrder] = useState<string[] | null>(null);
  const { mutate: reorder } = useReorderFavoritesMutation();
  const sensors = useSortSensors();
  const justDragged = useRef(false);

  if (isPending) {
    return (
      <div className="flex flex-col gap-2 px-2.5 py-1">
        <Skeleton className="h-3 w-2/3" />
      </div>
    );
  }
  if (error) {
    return (
      <Button
        variant="ghost"
        size="sm"
        onClick={() => void refetch()}
        className="text-destructive justify-start"
      >
        Could not load favorites. Retry
      </Button>
    );
  }
  const favorites = data.favorites;
  if (favorites.length === 0) {
    return (
      <p className="text-muted-foreground px-2.5 text-xs">
        Star issues, projects, cycles and views to keep them here.
      </p>
    );
  }
  const ordered = dragOrder
    ? dragOrder
        .map((id) => favorites.find((favorite) => favorite.id === id))
        .filter((favorite): favorite is TFavorite => !!favorite)
    : favorites;

  const onDragEnd = (event: DragEndEvent) => {
    justDragged.current = true;
    const moved = reorderedPosition(ordered, event);
    if (!moved) return;
    setDragOrder(moved.order);
    reorder(
      { workspaceId: workspace.id, favoriteIds: moved.order },
      { onSettled: () => setDragOrder(null) },
    );
  };
  const swallowClickAfterDrag = (event: MouseEvent) => {
    if (!justDragged.current) return;
    justDragged.current = false;
    event.preventDefault();
    event.stopPropagation();
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={() => (justDragged.current = false)}
      onDragEnd={onDragEnd}
    >
      <SortableContext
        items={ordered.map((favorite) => favorite.id)}
        strategy={verticalListSortingStrategy}
      >
        <div
          role="list"
          aria-label="Favorites"
          className="flex flex-col gap-0.5"
          onClickCapture={swallowClickAfterDrag}
        >
          {ordered.map((favorite) => (
            <FavoriteItem key={favorite.id} favorite={favorite} onNavigate={onNavigate} />
          ))}
        </div>
      </SortableContext>
    </DndContext>
  );
}

function FavoriteItem({ favorite, onNavigate }: { favorite: TFavorite } & TProps) {
  const { workspace } = useWorkspace();
  const sortable = useSortable({ id: favorite.id });
  const link = favoriteLink(workspace.slug, favorite);
  if (!link) return null;
  return (
    <div
      ref={sortable.setNodeRef}
      {...sortable.attributes}
      {...sortable.listeners}
      role="listitem"
      style={sortableStyle(sortable.transform, sortable.transition)}
      data-dragging={sortable.isDragging || undefined}
      className="data-dragging:z-10 data-dragging:opacity-80"
    >
      <SidebarNavItem
        label={favorite.name}
        icon={favoriteKindIcons[favorite.kind]}
        link={link}
        onNavigate={onNavigate}
      />
    </div>
  );
}
