import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import {
  ArchiveIcon,
  ArchiveRestoreIcon,
  EllipsisIcon,
  FileTextIcon,
  TargetIcon,
  Trash2Icon,
  UndoIcon,
  XIcon,
} from "lucide-react";
import { useState } from "react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { AutosaveMarkdown } from "@/components/projects/autosave-markdown";
import { InlineText } from "@/components/projects/inline-text";
import { LifecycleBanner } from "@/components/documents/lifecycle-banner";
import { useDocumentLifecycle } from "@/components/documents/use-document-lifecycle";
import { Breadcrumb, type TBreadcrumbItem } from "@/components/layout/breadcrumb";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { EntityIcon } from "@/components/shared/entity-icon";
import { IconPicker } from "@/components/shared/icon-picker";
import { ProjectIcon } from "@/components/shared/project-icon";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";
import {
  DOCUMENT_TITLE_MAX_LENGTH,
  documentBySlugQuery,
  type TDocument,
  useUpdateDocumentMutation,
} from "@/lib/queries/documents";

export function DocumentDetail({ slug }: { slug: string }) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(documentBySlugQuery(workspace.id, slug));
  if (isPending) return <ListSkeleton rows={4} />;
  if (error) {
    if (isApiError(error, 404)) {
      return (
        <EmptyState
          icon={FileTextIcon}
          title="Document not found"
          description="It was deleted, or you cannot see it."
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
        <FavoriteButton kind="document" entityId={document.id} />
        {canEdit && <DocumentActions document={document} />}
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
  const { workspace, isAdmin } = useWorkspace();
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
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon" aria-label="Document actions">
              <EllipsisIcon className="size-4" />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="w-48">
          <DropdownMenuItem onClick={() => run(document.archived_at ? "unarchive" : "archive")}>
            {document.archived_at ? (
              <ArchiveRestoreIcon className="size-4" />
            ) : (
              <ArchiveIcon className="size-4" />
            )}
            {document.archived_at ? "Unarchive" : "Archive"}
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => run(document.trashed_at ? "restore" : "trash")}>
            {document.trashed_at ? (
              <UndoIcon className="size-4" />
            ) : (
              <Trash2Icon className="size-4" />
            )}
            {document.trashed_at ? "Restore from trash" : "Move to trash"}
          </DropdownMenuItem>
          {isAdmin && (
            <>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-destructive" onClick={() => setIsDeleting(true)}>
                <XIcon className="size-4" />
                Delete permanently
              </DropdownMenuItem>
            </>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
      <ConfirmDialog
        open={isDeleting}
        onOpenChange={setIsDeleting}
        title="Delete document"
        description={`"${document.title}" is deleted for everyone. This cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={deleteDocument}
      />
    </>
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
