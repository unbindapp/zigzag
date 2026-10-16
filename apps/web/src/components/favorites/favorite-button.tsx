import { useQuery } from "@tanstack/react-query";
import { StarIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  favoritesListQuery,
  type TFavoriteKind,
  useAddFavoriteMutation,
  useRemoveFavoriteMutation,
} from "@/lib/queries/favorites";

type TProps = {
  kind: TFavoriteKind;
  entityId: string;
};

export function FavoriteButton({ kind, entityId }: TProps) {
  const { workspace } = useWorkspace();
  const { data } = useQuery(favoritesListQuery(workspace.id));
  const { mutate: add, isPending: isAdding } = useAddFavoriteMutation();
  const { mutate: remove, isPending: isRemoving } = useRemoveFavoriteMutation();
  const isFavorite = (data?.favorites ?? []).some(
    (favorite) => favorite.kind === kind && favorite.entity_id === entityId,
  );
  const label = isFavorite ? "Remove from favorites" : "Add to favorites";
  const target = { workspace_id: workspace.id, kind, entity_id: entityId };

  return (
    <SimpleTooltip content={label}>
      <Button
        variant="ghost"
        size="icon"
        aria-label={label}
        aria-pressed={isFavorite}
        disabled={isAdding || isRemoving}
        onClick={() => (isFavorite ? remove(target) : add(target))}
      >
        <StarIcon
          data-active={isFavorite || undefined}
          className="data-active:text-warning size-4 data-active:fill-current"
        />
      </Button>
    </SimpleTooltip>
  );
}
