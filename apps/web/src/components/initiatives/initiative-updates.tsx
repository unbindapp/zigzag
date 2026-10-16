import { useQuery } from "@tanstack/react-query";
import { Trash2Icon } from "lucide-react";

import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { ListStates } from "@/components/layout/list-states";
import { SectionTitle } from "@/components/layout/section-title";
import { StatusUpdateCard } from "@/components/projects/status-update-card";
import { StatusUpdateComposer } from "@/components/projects/status-update-composer";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  initiativeUpdatesQuery,
  type TInitiative,
  type TInitiativeUpdate,
  useCreateInitiativeUpdateMutation,
  useDeleteInitiativeUpdateMutation,
} from "@/lib/queries/initiatives";
import { meQuery } from "@/lib/queries/me";

function newestFirst(a: TInitiativeUpdate, b: TInitiativeUpdate) {
  return b.created_at.localeCompare(a.created_at);
}

export function InitiativeUpdates({ initiative }: { initiative: TInitiative }) {
  const { canEdit } = useWorkspace();
  const { users } = useWorkspaceUsers();
  const { data, isPending, error, refetch } = useQuery(initiativeUpdatesQuery(initiative.id));
  const { mutateAsync: createUpdate } = useCreateInitiativeUpdateMutation();
  const updates = [...(data?.updates ?? [])].sort(newestFirst);
  return (
    <section aria-label="Updates" className="flex w-full flex-col gap-4">
      <SectionTitle title="Updates" />
      <div className="flex w-full flex-col gap-3">
        {canEdit && (
          <StatusUpdateComposer
            users={users}
            initialHealth={initiative.health ?? "on_track"}
            onSubmit={(value) => createUpdate({ initiative_id: initiative.id, ...value })}
          />
        )}
        <ListStates
          isPending={isPending}
          error={error}
          onRetry={() => void refetch()}
          isEmpty={updates.length === 0}
          skeletonRows={2}
          empty={null}
        >
          {updates.map((update) => (
            <InitiativeUpdateCard key={update.id} update={update} />
          ))}
        </ListStates>
      </div>
    </section>
  );
}

function InitiativeUpdateCard({ update }: { update: TInitiativeUpdate }) {
  const { isAdmin, canEdit } = useWorkspace();
  const { data: me } = useQuery(meQuery);
  const { mutateAsync: deleteUpdate } = useDeleteInitiativeUpdateMutation();
  const isAuthor = Boolean(me && update.author?.id === me.user.id);
  const canDelete = canEdit && (isAuthor || isAdmin);
  return (
    <StatusUpdateCard
      author={update.author}
      body={update.body}
      health={update.health}
      createdAt={update.created_at}
      editedAt={update.edited_at}
      actions={
        canDelete && (
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label="Delete update">
                <Trash2Icon className="size-4" />
              </Button>
            }
            title="Delete update"
            description="The update is deleted for everyone."
            confirmLabel="Delete update"
            onConfirm={() => deleteUpdate({ updateId: update.id })}
          />
        )
      }
    />
  );
}
