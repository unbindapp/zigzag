import { useQuery } from "@tanstack/react-query";
import { ArchiveIcon, CircleDotIcon, Trash2Icon } from "lucide-react";
import { useRef, useState } from "react";

import { useCommandPanelStore } from "@/components/command-panel/command-panel-store";
import { useUploadImage } from "@/components/issues/comment-thread";
import { IssueActivity } from "@/components/issues/issue-activity";
import { IssueAttachments } from "@/components/issues/issue-attachments";
import { IssueDetailHeader } from "@/components/issues/issue-detail-header";
import { draftFromIssue } from "@/components/issues/issue-draft";
import { IssueProperties } from "@/components/issues/issue-properties";
import { IssueRelations } from "@/components/issues/issue-relations";
import { IssueTitle } from "@/components/issues/issue-title";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { SubIssues } from "@/components/issues/sub-issues";
import {
  type TIssueDetail,
  issueQuery,
  useIssueSubscriptionMutation,
  useUpdateIssueMutation,
} from "@/lib/queries/issues";
import { useIssueLifecycle, useIssuePatch } from "@/components/issues/use-issue-actions";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { MarkdownEditor } from "@/components/shared/markdown-editor";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";

type TProps = {
  identifier: string;
  variant?: "page" | "peek";
  onClose?: () => void;
};

export function IssueDetail({ identifier, variant = "page", onClose }: TProps) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(issueQuery(workspace.id, identifier));
  if (isPending) {
    return (
      <div className="flex w-full flex-col gap-6 p-4">
        <ListSkeleton rows={4} />
      </div>
    );
  }
  if (error) {
    if (isApiError(error, 404)) {
      return (
        <EmptyState
          icon={CircleDotIcon}
          title="Issue not found"
          description={`There is no issue ${identifier.toUpperCase()}, or you cannot see it.`}
        />
      );
    }
    return (
      <div className="p-4">
        <ErrorCard message={error.message} onRetry={() => void refetch()} />
      </div>
    );
  }
  return <IssueDetailContent issue={data.issue} variant={variant} onClose={onClose} />;
}

function IssueDetailContent({
  issue,
  variant,
  onClose,
}: {
  issue: TIssueDetail;
  variant: "page" | "peek";
  onClose?: () => void;
}) {
  const { canEdit } = useWorkspace();
  const patch = useIssuePatch();
  const { mutateAsync: updateIssue } = useUpdateIssueMutation();
  const { mutate: setSubscribed } = useIssueSubscriptionMutation();
  const isDialogOpen = useNewIssueStore((state) => state.isOpen);
  const isCommandPanelOpen = useCommandPanelStore((state) => state.isOpen);

  const properties = (
    <IssueProperties
      teamId={issue.team_id}
      issueId={issue.id}
      value={draftFromIssue(issue)}
      onChange={(next) => patch(issue, next)}
      disabled={!canEdit}
      shortcutsEnabled={!isDialogOpen && !isCommandPanelOpen}
      subscribers={{
        users: issue.subscribers ?? [],
        onChange: (added, removed) => {
          added.forEach((user) =>
            setSubscribed({ issueId: issue.id, userId: user.id, subscribed: true }),
          );
          removed.forEach((user) =>
            setSubscribed({ issueId: issue.id, userId: user.id, subscribed: false }),
          );
        },
      }}
    />
  );

  const main = (
    <div className="flex min-w-0 flex-1 flex-col gap-6">
      <LifecycleBanner issue={issue} />
      <div className="flex flex-col gap-2">
        <IssueTitle
          title={issue.title}
          readOnly={!canEdit}
          onSave={(title) =>
            void updateIssue({ input: { issue_id: issue.id, title }, optimistic: { title } }).catch(
              () => undefined,
            )
          }
        />
        <DescriptionEditor
          key={issue.id}
          initialValue={issue.description}
          readOnly={!canEdit}
          onSave={(description) => updateIssue({ input: { issue_id: issue.id, description } })}
        />
      </div>
      <SubIssues issue={issue} />
      <IssueRelations issue={issue} />
      <IssueAttachments issue={issue} />
      <IssueActivity issueId={issue.id} />
    </div>
  );

  if (variant === "peek") {
    return (
      <div className="flex w-full flex-col gap-6 p-4">
        <IssueDetailHeader issue={issue} variant="peek" onClose={onClose} />
        <div className="rounded-lg border p-4">{properties}</div>
        {main}
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-6">
      <IssueDetailHeader issue={issue} variant="page" />
      <div className="flex w-full flex-col gap-6 lg:flex-row">
        {main}
        <aside aria-label="Properties" className="w-full shrink-0 lg:w-72">
          <div className="rounded-lg border p-4 lg:sticky lg:top-0">{properties}</div>
        </aside>
      </div>
    </div>
  );
}

function DescriptionEditor({
  initialValue,
  readOnly,
  onSave,
}: {
  initialValue: string;
  readOnly: boolean;
  onSave: (description: string) => Promise<unknown>;
}) {
  const [value] = useState(initialValue);
  const lastSaved = useRef(initialValue);
  const { users } = useWorkspaceUsers();
  const uploadImage = useUploadImage();
  return (
    <MarkdownEditor
      label="Description"
      value={value}
      placeholder="Add a description…"
      readOnly={readOnly}
      users={users}
      onUploadImage={uploadImage}
      autosave={{
        onSave: async (description) => {
          if (description === lastSaved.current) return;
          await onSave(description);
          lastSaved.current = description;
        },
      }}
    />
  );
}

function LifecycleBanner({ issue }: { issue: TIssueDetail }) {
  const { canEdit } = useWorkspace();
  const lifecycle = useIssueLifecycle();
  if (!issue.trashed_at && !issue.archived_at) return null;
  const isTrashed = Boolean(issue.trashed_at);
  const Icon = isTrashed ? Trash2Icon : ArchiveIcon;
  return (
    <div role="status" className="bg-muted flex items-center gap-2 rounded-lg border p-4 text-sm">
      <Icon className="text-muted-foreground size-4 shrink-0" />
      <p className="min-w-0 flex-1">
        {isTrashed
          ? "This issue is in the trash and is deleted for good 30 days after it was trashed."
          : "This issue is archived."}
      </p>
      {canEdit && (
        <Button
          variant="outline"
          size="sm"
          isPending={lifecycle.isPending}
          onClick={() =>
            void lifecycle
              .run([issue.id], isTrashed ? "restore" : "unarchive")
              .catch(() => undefined)
          }
        >
          {isTrashed ? "Restore" : "Unarchive"}
        </Button>
      )}
    </div>
  );
}
