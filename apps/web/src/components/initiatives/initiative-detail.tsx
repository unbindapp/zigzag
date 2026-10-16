import { useQuery } from "@tanstack/react-query";
import { TargetIcon } from "lucide-react";

import { AutosaveMarkdown } from "@/components/projects/autosave-markdown";
import { InlineText } from "@/components/projects/inline-text";
import { LifecycleBanner } from "@/components/documents/lifecycle-banner";
import { ProjectDocuments } from "@/components/documents/project-documents";
import {
  InitiativeActions,
  useInitiativeLifecycle,
} from "@/components/initiatives/initiative-actions";
import { InitiativeProjects } from "@/components/initiatives/initiative-projects";
import { InitiativeProperties } from "@/components/initiatives/initiative-properties";
import { InitiativeUpdates } from "@/components/initiatives/initiative-updates";
import { Breadcrumb } from "@/components/layout/breadcrumb";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { IconPicker } from "@/components/shared/icon-picker";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";
import {
  initiativeBySlugQuery,
  initiativeNameMaxLength,
  type TInitiative,
  useUpdateInitiativeMutation,
} from "@/lib/queries/initiatives";

export function InitiativeDetail({ slug }: { slug: string }) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(initiativeBySlugQuery(workspace.id, slug));
  if (isPending) return <ListSkeleton rows={4} />;
  if (error) {
    if (isApiError(error, 404)) {
      return (
        <EmptyState
          icon={TargetIcon}
          title="Initiative not found"
          description="It was deleted, or you cannot see it."
        />
      );
    }
    return <ErrorCard message={error.message} onRetry={() => void refetch()} />;
  }
  return <InitiativeDetailContent key={data.initiative.id} initiative={data.initiative} />;
}

function InitiativeDetailContent({ initiative }: { initiative: TInitiative }) {
  const { workspace, canEdit } = useWorkspace();
  const { mutate: update, mutateAsync: updateAsync } = useUpdateInitiativeMutation();
  const lifecycle = useInitiativeLifecycle();

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex w-full min-w-0 items-center gap-2">
        <div className="min-w-0 flex-1">
          <Breadcrumb
            items={[
              {
                label: "Initiatives",
                link: { to: "/$workspace/initiatives", params: { workspace: workspace.slug } },
              },
              { label: initiative.name },
            ]}
          />
        </div>
        <InitiativeActions initiative={initiative} />
      </div>
      <div className="flex w-full flex-col gap-6 lg:flex-row">
        <div className="flex min-w-0 flex-1 flex-col gap-6">
          <LifecycleBanner
            entity="initiative"
            state={initiative.archived_at ? "archived" : null}
            isPending={lifecycle.isPending}
            onRestore={
              canEdit
                ? () => void lifecycle.run(initiative.id, "unarchive").catch(() => undefined)
                : undefined
            }
          />
          <div className="flex flex-col gap-2">
            <div className="flex min-w-0 items-center gap-2">
              <IconPicker
                value={initiative.icon || null}
                color={initiative.color || null}
                disabled={!canEdit}
                label="Initiative icon"
                onChange={(icon) => update({ initiative_id: initiative.id, icon: icon ?? "" })}
              />
              <InlineText
                value={initiative.name}
                label="Initiative name"
                maxLength={initiativeNameMaxLength}
                readOnly={!canEdit}
                onSave={(name) => update({ initiative_id: initiative.id, name })}
              />
            </div>
            <AutosaveMarkdown
              key={initiative.id}
              label="Description"
              placeholder="Describe the initiative…"
              initialValue={initiative.description}
              readOnly={!canEdit}
              onSave={(description) => updateAsync({ initiative_id: initiative.id, description })}
            />
          </div>
          <InitiativeProjects initiative={initiative} />
          <InitiativeUpdates initiative={initiative} />
          <ProjectDocuments workspaceId={workspace.id} initiativeId={initiative.id} />
        </div>
        <aside aria-label="Properties" className="w-full shrink-0 lg:w-72">
          <div className="rounded-lg border p-4 lg:sticky lg:top-0">
            <InitiativeProperties initiative={initiative} />
          </div>
        </aside>
      </div>
    </div>
  );
}
