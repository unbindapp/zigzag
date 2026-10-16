import { useQuery } from "@tanstack/react-query";
import { BotIcon } from "lucide-react";

import { describeChange } from "@/components/issues/activity-text";
import { CommentComposer, CommentThread } from "@/components/issues/comment-thread";
import { UserAvatar } from "@/components/issues/pickers";
import { type TComment, type THistoryEntry, issueHistoryQuery } from "@/lib/queries/issues";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SectionTitle } from "@/components/layout/section-title";
import { RelativeTime } from "@/components/shared/relative-time";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { commentsQuery } from "@/lib/queries/comments";

type TFeedItem =
  | { kind: "history"; id: string; at: string; entry: THistoryEntry }
  | { kind: "comment"; id: string; at: string; comment: TComment };

export function IssueActivity({ issueId }: { issueId: string }) {
  const { canEdit } = useWorkspace();
  const history = useQuery(issueHistoryQuery(issueId));
  const comments = useQuery(commentsQuery(issueId));

  const items: TFeedItem[] = [
    ...(history.data?.history ?? []).map((entry): TFeedItem => ({
      kind: "history",
      id: entry.id,
      at: entry.created_at,
      entry,
    })),
    ...(comments.data?.comments ?? []).map((comment): TFeedItem => ({
      kind: "comment",
      id: comment.id,
      at: comment.created_at,
      comment,
    })),
  ].sort((a, b) => a.at.localeCompare(b.at));

  const error = history.error ?? comments.error;
  return (
    <section aria-label="Activity" className="flex w-full flex-col gap-4">
      <SectionTitle title="Activity" />
      {history.isPending || comments.isPending ? (
        <ListSkeleton rows={3} />
      ) : error ? (
        <ErrorCard
          message={error.message}
          onRetry={() => {
            void history.refetch();
            void comments.refetch();
          }}
        />
      ) : (
        <ol className="flex flex-col gap-3">
          {items.map((item) => (
            <li key={`${item.kind}-${item.id}`} className="flex flex-col">
              {item.kind === "history" ? (
                <HistoryLine entry={item.entry} />
              ) : (
                <CommentThread comment={item.comment} issueId={issueId} />
              )}
            </li>
          ))}
        </ol>
      )}
      {canEdit && <CommentComposer issueId={issueId} />}
    </section>
  );
}

function HistoryLine({ entry }: { entry: THistoryEntry }) {
  const actorName = entry.actor ? entry.actor.name || entry.actor.display_name : "Zigzag";
  return (
    <div className="flex flex-col gap-1">
      {entry.changes.map((change, index) => (
        <p
          key={`${change.field}-${index}`}
          className="text-muted-foreground flex min-w-0 items-center gap-1.5 px-1 text-xs"
        >
          {entry.actor ? (
            <UserAvatar user={entry.actor} size="xs" />
          ) : (
            <BotIcon className="size-3.5 shrink-0" />
          )}
          <span className="text-foreground shrink-0 font-medium">{actorName}</span>
          <span className="min-w-0 truncate">{describeChange(change)}</span>
          <span aria-hidden>·</span>
          <span className="shrink-0">
            <RelativeTime value={entry.created_at} />
          </span>
        </p>
      ))}
    </div>
  );
}
