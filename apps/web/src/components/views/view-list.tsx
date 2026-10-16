import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { LayersIcon, LockIcon, UsersIcon } from "lucide-react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { EntityIcon } from "@/components/shared/entity-icon";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useCanEditView, ViewMenu } from "@/components/views/view-menu";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TView, viewsListQuery } from "@/lib/queries/views";

export function ViewList() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(viewsListQuery(workspace.id));
  const views = data?.views ?? [];
  return (
    <ListStates
      isPending={isPending}
      error={error}
      onRetry={() => void refetch()}
      isEmpty={views.length === 0}
      empty={{
        icon: LayersIcon,
        title: "No views yet",
        description: "Filter an issue list, then choose Save view to keep the filters here.",
      }}
    >
      <List>
        {views.map((view) => (
          <ViewRow key={view.id} view={view} />
        ))}
      </List>
    </ListStates>
  );
}

function ViewRow({ view }: { view: TView }) {
  const { workspace } = useWorkspace();
  const canEdit = useCanEditView(view);
  return (
    <div
      role="listitem"
      className="flex h-9 w-full items-center gap-2 border-b px-3 text-sm last:border-b-0"
    >
      <EntityIcon icon={view.icon} color={view.color} fallback={LayersIcon} />
      <Link
        to="/$workspace/view/$view_id"
        params={{ workspace: workspace.slug, view_id: view.id }}
        className="min-w-0 truncate hover:underline"
      >
        {view.name}
      </Link>
      {view.description && (
        <span className="text-muted-foreground hidden min-w-0 truncate text-xs sm:inline">
          {view.description}
        </span>
      )}
      <span className="flex-1" />
      {!view.is_owner && view.owner && (
        <span className="text-muted-foreground hidden text-xs md:inline">{view.owner.name}</span>
      )}
      <SimpleTooltip content={view.shared ? "Shared with the workspace" : "Only you"}>
        {view.shared ? (
          <UsersIcon aria-label="Shared" className="text-muted-foreground size-4" />
        ) : (
          <LockIcon aria-label="Private" className="text-muted-foreground size-4" />
        )}
      </SimpleTooltip>
      <FavoriteButton kind="view" entityId={view.id} />
      {canEdit && <ViewMenu view={view} />}
    </div>
  );
}
