import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { ArchiveIcon, CircleDotIcon, Trash2Icon } from "lucide-react";
import { useEffect } from "react";

import { IssueActivity } from "@/components/issues/issue-activity";
import { IssueAttachments } from "@/components/issues/issue-attachments";
import { IssueDetailHeader } from "@/components/issues/issue-detail-header";
import { draftFromIssue } from "@/components/issues/issue-draft";
import { IssueProperties } from "@/components/issues/issue-properties";
import { IssueRelations } from "@/components/issues/issue-relations";
import { SubIssues } from "@/components/issues/sub-issues";
import {
  ISSUE_TITLE_MAX_LENGTH,
  type TIssueDetail,
  issueQuery,
  useIssueSubscriptionMutation,
  useUpdateIssueMutation,
} from "@/lib/queries/issues";
import { useIssueLifecycle, useIssuePatch } from "@/components/issues/use-issue-actions";
import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { AutosaveMarkdown } from "@/components/shared/autosave-markdown";
import { InlineText } from "@/components/shared/inline-text";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";
import { EntityNotFound } from "@/components/navigation/not-found-template";

type TProps = {
  identifier: string;
  variant?: "page" | "peek";
  onClose?: () => void;
  shortcuts?: boolean;
};

export function IssueDetail({ identifier, variant = "page", onClose, shortcuts = true }: TProps) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(issueQuery(workspace.id, identifier));
  const navigate = useNavigate();
  const canonical = data?.issue.identifier;
  useEffect(() => {
    if (variant !== "page" || !canonical || canonical === identifier) return;
    void navigate({
      to: "/$workspace/issue/$identifier",
      params: { workspace: workspace.slug, identifier: canonical },
      replace: true,
    });
  }, [variant, canonical, identifier, navigate, workspace.slug]);
  if (isPending) {
    return (
      <div className="flex w-full flex-col gap-6 p-4">
        <ListSkeleton rows={4} />
      </div>
    );
  }
  if (error) {
    if (isApiError(error, 404) && variant === "page") {
      return (
        <EntityNotFound
          Icon={CircleDotIcon}
          description={`There is no issue ${identifier.toUpperCase()}, or you cannot see it.`}
        />
      );
    }
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
  return (
    <IssueDetailContent
      issue={data.issue}
      variant={variant}
      onClose={onClose}
      shortcuts={shortcuts}
    />
  );
}

function IssueDetailContent({
  issue,
  variant,
  onClose,
  shortcuts,
}: {
  issue: TIssueDetail;
  variant: "page" | "peek";
  onClose?: () => void;
  shortcuts: boolean;
}) {
  const { canEdit } = useWorkspace();
  const patch = useIssuePatch();
  const { mutateAsync: updateIssue } = useUpdateIssueMutation();
  const { mutate: setSubscribed } = useIssueSubscriptionMutation();

  const properties = (
    <IssueProperties
      teamId={issue.team_id}
      issueId={issue.id}
      value={draftFromIssue(issue)}
      onChange={(next) => patch(issue, next)}
      disabled={!canEdit}
      shortcutsEnabled={shortcuts}
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
        <InlineText
          label="Issue title"
          value={issue.title}
          maxLength={ISSUE_TITLE_MAX_LENGTH}
          readOnly={!canEdit}
          onSave={(title) =>
            void updateIssue({ input: { issue_id: issue.id, title }, optimistic: { title } }).catch(
              () => undefined,
            )
          }
        />
        <AutosaveMarkdown
          key={issue.id}
          label="Description"
          placeholder="Add a description…"
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
