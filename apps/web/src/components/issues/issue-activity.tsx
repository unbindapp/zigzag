import { useInfiniteQuery } from "@tanstack/react-query";
import { BotIcon } from "lucide-react";

import { mergeFeed } from "@/components/issues/activity-feed";
import {
  describeChange,
  historyActorOf,
  readsWithoutActor,
  type THistoryActor,
} from "@/components/issues/activity-text";
import { GithubIcon } from "@/components/icons/github";
import { CommentComposer, CommentThread } from "@/components/issues/comment-thread";
import { UserAvatar } from "@/components/issues/pickers";
import { type TComment, type THistoryEntry, issueHistoryQuery } from "@/lib/queries/issues";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SectionTitle } from "@/components/layout/section-title";
import { RelativeTime } from "@/components/shared/relative-time";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { commentsQuery } from "@/lib/queries/comments";

type TFeedItem =
  | { kind: "history"; id: string; at: string; entry: THistoryEntry }
  | { kind: "comment"; id: string; at: string; comment: TComment };

export function IssueActivity({ issueId }: { issueId: string }) {
  const { canEdit } = useWorkspace();
  const history = useInfiniteQuery(issueHistoryQuery(issueId));
  const comments = useInfiniteQuery(commentsQuery(issueId));

  const items = mergeFeed<TFeedItem>([
    {
      items: (history.data?.pages ?? []).flatMap((page) =>
        page.data.map((entry): TFeedItem => ({
          kind: "history",
          id: entry.id,
          at: entry.created_at,
          entry,
        })),
      ),
      hasMore: history.hasNextPage,
    },
    {
      items: (comments.data?.pages ?? []).flatMap((page) =>
        page.data.map((comment): TFeedItem => ({
          kind: "comment",
          id: comment.id,
          at: comment.created_at,
          comment,
        })),
      ),
      hasMore: comments.hasNextPage,
    },
  ]);
  const hasMore = history.hasNextPage || comments.hasNextPage;
  const isFetchingMore = history.isFetchingNextPage || comments.isFetchingNextPage;
  const loadMore = () => {
    if (history.hasNextPage) void history.fetchNextPage();
    if (comments.hasNextPage) void comments.fetchNextPage();
  };

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
      {hasMore && (
        <div className="flex">
          <Button variant="outline" size="sm" isPending={isFetchingMore} onClick={loadMore}>
            Load more
          </Button>
        </div>
      )}
      {canEdit && <CommentComposer issueId={issueId} />}
    </section>
  );
}

function HistoryLine({ entry }: { entry: THistoryEntry }) {
  const actor = historyActorOf(entry);
  return (
    <div className="flex flex-col gap-1">
      {entry.changes.map((change, index) => (
        <p
          key={`${change.field}-${index}`}
          className="text-muted-foreground flex min-w-0 items-center gap-1.5 px-1 text-xs"
        >
          {readsWithoutActor(change) ? (
            <GithubIcon className="size-3.5" />
          ) : (
            <ActorMark entry={entry} actor={actor} />
          )}
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

function ActorMark({ entry, actor }: { entry: THistoryEntry; actor: THistoryActor }) {
  switch (actor.kind) {
    case "user":
      return (
        <>
          {entry.actor && <UserAvatar user={entry.actor} size="xs" />}
          <span className="text-foreground shrink-0 font-medium">{actor.name}</span>
        </>
      );
    case "github":
      return (
        <>
          <GithubIcon className="size-3.5" />
          <span className="text-foreground shrink-0 font-medium">GitHub</span>
        </>
      );
    default:
      return (
        <>
          <BotIcon className="size-3.5 shrink-0" />
          <span className="text-foreground shrink-0 font-medium">Zigzag</span>
        </>
      );
  }
}
