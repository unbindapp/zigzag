import { useNavigate } from "@tanstack/react-router";
import { ArchiveIcon, ArchiveRestoreIcon, Trash2Icon } from "lucide-react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  type TInitiative,
  type TInitiativeLifecycleAction,
  useInitiativeLifecycleMutation,
} from "@/lib/queries/initiatives";

const lifecycleToasts: Record<TInitiativeLifecycleAction, string> = {
  archive: "Initiative archived",
  unarchive: "Initiative unarchived",
  delete: "Initiative deleted",
};

export function useInitiativeLifecycle() {
  const { mutateAsync, isPending } = useInitiativeLifecycleMutation();
  const run = async (initiativeId: string, action: TInitiativeLifecycleAction) => {
    await mutateAsync({ initiativeId, action });
    toast.add({ type: "success", title: lifecycleToasts[action] });
  };
  return { run, isPending };
}

export function InitiativeActions({ initiative }: { initiative: TInitiative }) {
  const { workspace, canEdit, isAdmin } = useWorkspace();
  const navigate = useNavigate();
  const lifecycle = useInitiativeLifecycle();
  const archiveLabel = initiative.archived_at ? "Unarchive" : "Archive";

  const deleteInitiative = async () => {
    await lifecycle.run(initiative.id, "delete");
    await navigate({ to: "/$workspace/initiatives", params: { workspace: workspace.slug } });
  };

  return (
    <div className="flex shrink-0 items-center gap-1">
      <FavoriteButton kind="initiative" entityId={initiative.id} />
      {canEdit && (
        <SimpleTooltip content={archiveLabel}>
          <Button
            variant="ghost"
            size="icon"
            aria-label={archiveLabel}
            isPending={lifecycle.isPending}
            onClick={() =>
              void lifecycle
                .run(initiative.id, initiative.archived_at ? "unarchive" : "archive")
                .catch(() => undefined)
            }
          >
            {initiative.archived_at ? (
              <ArchiveRestoreIcon className="size-4" />
            ) : (
              <ArchiveIcon className="size-4" />
            )}
          </Button>
        </SimpleTooltip>
      )}
      {isAdmin && (
        <ConfirmDialog
          trigger={
            <Button variant="ghost" size="icon" aria-label="Delete initiative">
              <Trash2Icon className="size-4" />
            </Button>
          }
          title="Delete initiative"
          description={`"${initiative.name}" is deleted for everyone. Its projects stay, without this initiative.`}
          confirmLabel="Delete"
          onConfirm={deleteInitiative}
        />
      )}
    </div>
  );
}
