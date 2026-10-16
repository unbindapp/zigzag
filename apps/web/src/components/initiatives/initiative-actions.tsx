import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import {
  ArchiveAction,
  CopyLinkAction,
  DeleteMenuAction,
  HeaderActions,
} from "@/components/shared/detail-actions";
import { toast } from "@/components/ui/toast";
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
  const [isDeleting, setIsDeleting] = useState(false);

  const deleteInitiative = async () => {
    await lifecycle.run(initiative.id, "delete");
    await navigate({ to: "/$workspace/initiatives", params: { workspace: workspace.slug } });
  };

  return (
    <HeaderActions>
      <FavoriteButton kind="initiative" entityId={initiative.id} />
      <CopyLinkAction url={window.location.href} />
      {canEdit && (
        <ArchiveAction
          isOn={Boolean(initiative.archived_at)}
          isPending={lifecycle.isPending}
          onToggle={() =>
            void lifecycle
              .run(initiative.id, initiative.archived_at ? "unarchive" : "archive")
              .catch(() => undefined)
          }
        />
      )}
      {isAdmin && (
        <>
          <DeleteMenuAction label="Delete initiative" onSelect={() => setIsDeleting(true)} />
          <ConfirmDialog
            open={isDeleting}
            onOpenChange={setIsDeleting}
            title={`Delete ${initiative.name}`}
            description="The initiative is deleted for everyone. Its projects stay, without this initiative."
            confirmLabel="Delete initiative"
            onConfirm={deleteInitiative}
          />
        </>
      )}
    </HeaderActions>
  );
}
