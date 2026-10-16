import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { FileTextIcon, TargetIcon } from "lucide-react";
import { useState } from "react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { AutosaveMarkdown } from "@/components/shared/autosave-markdown";
import { InlineText } from "@/components/shared/inline-text";
import { LifecycleBanner } from "@/components/documents/lifecycle-banner";
import { useDocumentLifecycle } from "@/components/documents/use-document-lifecycle";
import { Breadcrumb, type TBreadcrumbItem } from "@/components/layout/breadcrumb";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import {
  ArchiveAction,
  CopyLinkAction,
  DeleteMenuAction,
  HeaderActions,
  TrashAction,
} from "@/components/shared/detail-actions";
import { EntityIcon } from "@/components/shared/entity-icon";
import { IconPicker } from "@/components/shared/icon-picker";
import { ProjectIcon } from "@/components/shared/project-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";
import {
  DOCUMENT_TITLE_MAX_LENGTH,
  documentBySlugQuery,
  type TDocument,
  useUpdateDocumentMutation,
} from "@/lib/queries/documents";
import { EntityNotFound } from "@/components/navigation/not-found-template";

export function DocumentDetail({ slug }: { slug: string }) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(documentBySlugQuery(workspace.id, slug));
  if (isPending) return <ListSkeleton rows={4} />;
  if (error) {
    if (isApiError(error, 404)) {
      return (
        <EntityNotFound
          Icon={FileTextIcon}
          description="Document not found. It was deleted, or you cannot see it."
        />
      );
    }
    return <ErrorCard message={error.message} onRetry={() => void refetch()} />;
  }
  return <DocumentDetailContent key={data.document.id} document={data.document} />;
}

function useDocumentBreadcrumb(document: TDocument): TBreadcrumbItem[] {
  const { workspace } = useWorkspace();
  const current = { label: document.title };
  if (document.project) {
    return [
      {
        label: document.project.name,
        icon: <ProjectIcon icon={document.project.icon} color={document.project.color} size="sm" />,
        link: {
          to: "/$workspace/project/$slug",
          params: { workspace: workspace.slug, slug: document.project.slug },
        },
      },
      current,
    ];
  }
  if (document.initiative) {
    return [
      {
        label: document.initiative.name,
        icon: (
          <EntityIcon
            icon={document.initiative.icon}
            color={document.initiative.color}
            fallback={TargetIcon}
            size="sm"
          />
        ),
        link: {
          to: "/$workspace/initiative/$slug",
          params: { workspace: workspace.slug, slug: document.initiative.slug },
        },
      },
      current,
    ];
  }
  return [
    {
      label: "Documents",
      link: { to: "/$workspace/documents", params: { workspace: workspace.slug } },
    },
    current,
  ];
}

function DocumentDetailContent({ document }: { document: TDocument }) {
  const { canEdit } = useWorkspace();
  const { mutate: update, mutateAsync: updateAsync } = useUpdateDocumentMutation();
  const lifecycle = useDocumentLifecycle();
  const breadcrumb = useDocumentBreadcrumb(document);
  const lifecycleState = document.trashed_at ? "trashed" : document.archived_at ? "archived" : null;

  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex w-full min-w-0 items-center gap-2">
        <div className="min-w-0 flex-1">
          <Breadcrumb items={breadcrumb} />
        </div>
        <DocumentActions document={document} />
      </div>
      <LifecycleBanner
        entity="document"
        state={lifecycleState}
        isPending={lifecycle.isPending}
        onRestore={
          canEdit
            ? () =>
                void lifecycle
                  .run(document.id, lifecycleState === "trashed" ? "restore" : "unarchive")
                  .catch(() => undefined)
            : undefined
        }
      />
      <div className="flex min-w-0 items-center gap-2">
        <IconPicker
          value={document.icon || null}
          disabled={!canEdit}
          label="Document icon"
          onChange={(icon) => update({ document_id: document.id, icon: icon ?? "" })}
        />
        <InlineText
          value={document.title}
          label="Document title"
          maxLength={DOCUMENT_TITLE_MAX_LENGTH}
          readOnly={!canEdit}
          onSave={(title) => update({ document_id: document.id, title })}
        />
      </div>
      <AutosaveMarkdown
        key={document.id}
        label="Document content"
        placeholder="Start writing…"
        initialValue={document.content}
        readOnly={!canEdit}
        onSave={(content) => updateAsync({ document_id: document.id, content })}
      />
    </div>
  );
}

function DocumentActions({ document }: { document: TDocument }) {
  const { workspace, canEdit, isAdmin } = useWorkspace();
  const navigate = useNavigate();
  const lifecycle = useDocumentLifecycle();
  const [isDeleting, setIsDeleting] = useState(false);
  const run = (action: "archive" | "unarchive" | "trash" | "restore") =>
    void lifecycle.run(document.id, action).catch(() => undefined);

  const deleteDocument = async () => {
    await lifecycle.run(document.id, "delete");
    await navigate(parentRoute(workspace.slug, document));
  };

  return (
    <HeaderActions>
      <FavoriteButton kind="document" entityId={document.id} />
      <CopyLinkAction url={window.location.href} />
      {canEdit && (
        <ArchiveAction
          isOn={Boolean(document.archived_at)}
          isPending={lifecycle.isPending}
          onToggle={() => run(document.archived_at ? "unarchive" : "archive")}
        />
      )}
      {canEdit && (
        <TrashAction
          isOn={Boolean(document.trashed_at)}
          isPending={lifecycle.isPending}
          onToggle={() => run(document.trashed_at ? "restore" : "trash")}
        />
      )}
      {isAdmin && (
        <>
          <DeleteMenuAction label="Delete document" onSelect={() => setIsDeleting(true)} />
          <ConfirmDialog
            open={isDeleting}
            onOpenChange={setIsDeleting}
            title={`Delete ${document.title}`}
            description="The document is deleted for everyone. This cannot be undone."
            confirmLabel="Delete document"
            onConfirm={deleteDocument}
          />
        </>
      )}
    </HeaderActions>
  );
}

function parentRoute(workspaceSlug: string, document: TDocument) {
  if (document.project) {
    return {
      to: "/$workspace/project/$slug",
      params: { workspace: workspaceSlug, slug: document.project.slug },
    } as const;
  }
  if (document.initiative) {
    return {
      to: "/$workspace/initiative/$slug",
      params: { workspace: workspaceSlug, slug: document.initiative.slug },
    } as const;
  }
  return { to: "/$workspace/documents", params: { workspace: workspaceSlug } } as const;
}
