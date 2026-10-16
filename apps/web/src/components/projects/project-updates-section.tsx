import { useQuery } from "@tanstack/react-query";
import { PencilIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { ReactionBar } from "@/components/issues/reaction-bar";
import { ListStates } from "@/components/layout/list-states";
import { SectionTitle } from "@/components/layout/section-title";
import { StatusUpdateCard } from "@/components/projects/status-update-card";
import { StatusUpdateComposer } from "@/components/projects/status-update-composer";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { meQuery } from "@/lib/queries/me";
import {
  type TProject,
  type TProjectUpdate,
  projectUpdatesQuery,
  useCreateProjectUpdateMutation,
  useDeleteProjectUpdateMutation,
  useEditProjectUpdateMutation,
  useProjectUpdateReactionMutation,
} from "@/lib/queries/projects";

export function ProjectUpdatesSection({ project }: { project: TProject }) {
  const { canEdit } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(projectUpdatesQuery(project.id));
  const { mutateAsync: createUpdate } = useCreateProjectUpdateMutation();
  const { users } = useWorkspaceUsers();
  const updates = data?.updates ?? [];

  return (
    <section aria-label="Updates" className="flex w-full flex-col gap-4">
      <SectionTitle title="Updates" />
      {canEdit && (
        <StatusUpdateComposer
          users={users}
          onSubmit={({ body, health }) => createUpdate({ project_id: project.id, body, health })}
        />
      )}
      <ListStates
        isPending={isPending}
        error={error}
        isEmpty={updates.length === 0}
        onRetry={() => void refetch()}
        empty={null}
      >
        <div className="flex flex-col gap-3">
          {updates.map((update) => (
            <ProjectUpdateItem key={update.id} projectId={project.id} update={update} />
          ))}
        </div>
      </ListStates>
    </section>
  );
}

function ProjectUpdateItem({ projectId, update }: { projectId: string; update: TProjectUpdate }) {
  const { canEdit, isAdmin } = useWorkspace();
  const { data: me } = useQuery(meQuery);
  const [isEditing, setIsEditing] = useState(false);
  const { mutateAsync: editUpdate } = useEditProjectUpdateMutation();
  const { mutateAsync: deleteUpdate } = useDeleteProjectUpdateMutation();
  const { mutate: toggleReaction } = useProjectUpdateReactionMutation(projectId);
  const isAuthor = Boolean(me && update.author?.id === me.user.id);
  const canManage = canEdit && (isAuthor || isAdmin);

  return (
    <StatusUpdateCard
      author={update.author}
      body={update.body}
      health={update.health}
      createdAt={update.created_at}
      editedAt={update.edited_at}
      actions={
        canManage &&
        !isEditing && (
          <>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Edit update"
              onClick={() => setIsEditing(true)}
            >
              <PencilIcon className="size-4" />
            </Button>
            <ConfirmDialog
              trigger={
                <Button variant="ghost" size="icon" aria-label="Delete update">
                  <Trash2Icon className="size-4" />
                </Button>
              }
              title="Delete update"
              description="The project's health falls back to the previous update."
              confirmLabel="Delete update"
              onConfirm={() => deleteUpdate({ updateId: update.id })}
            />
          </>
        )
      }
      footer={
        <ReactionBar
          reactions={update.reactions}
          disabled={!canEdit}
          onToggle={(emoji, reacted) => toggleReaction({ updateId: update.id, emoji, reacted })}
        />
      }
    >
      {isEditing ? (
        <StatusUpdateComposer
          initialBody={update.body}
          initialHealth={update.health}
          submitLabel="Save"
          onCancel={() => setIsEditing(false)}
          onSubmit={async ({ body, health }) => {
            await editUpdate({ update_id: update.id, body, health });
            setIsEditing(false);
          }}
        />
      ) : undefined}
    </StatusUpdateCard>
  );
}
