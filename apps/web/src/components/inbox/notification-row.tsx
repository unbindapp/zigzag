import { ArchiveIcon, ArchiveRestoreIcon, BellIcon } from "lucide-react";

import { notificationSummary, notificationTitle } from "@/components/inbox/notification-text";
import { Avatar } from "@/components/shared/avatar";
import { RelativeTime } from "@/components/shared/relative-time";
import { Button } from "@/components/ui/button";
import { cn } from "@/components/ui/utils";
import type { TNotification } from "@/lib/queries/notifications";

type TProps = {
  notification: TNotification;
  isActive: boolean;
  isArchived: boolean;
  onOpen: (notification: TNotification) => void;
  onToggleArchive: (notification: TNotification) => void;
};

export function NotificationRow({
  notification,
  isActive,
  isArchived,
  onOpen,
  onToggleArchive,
}: TProps) {
  const isUnread = !notification.read_at;
  const title = notificationTitle(notification);
  const actorName = notification.actor?.name || notification.actor?.display_name;
  const identifier = notification.issue?.identifier;
  const snoozedUntil = notification.snoozed_until;
  const isSnoozed = snoozedUntil !== null && new Date(snoozedUntil).getTime() > Date.now();

  return (
    <div
      role="listitem"
      tabIndex={0}
      aria-current={isActive || undefined}
      aria-label={`${isUnread ? "Unread: " : ""}${notification.title}`}
      data-active={isActive || undefined}
      data-unread={isUnread || undefined}
      onClick={() => onOpen(notification)}
      onKeyDown={(event) => {
        if (event.key !== "Enter") return;
        onOpen(notification);
      }}
      className="group/row has-hover:hover:bg-accent data-active:bg-accent focus-visible:ring-primary/8-10 flex h-9 w-full cursor-default items-center gap-2 border-b px-3 text-sm outline-hidden last:border-b-0 focus-visible:ring-1 focus-visible:ring-inset"
    >
      <span
        aria-hidden
        className={cn("size-2 shrink-0 rounded-full", isUnread ? "bg-primary" : "bg-transparent")}
      />
      {notification.actor ? (
        <Avatar
          id={notification.actor.id}
          name={actorName ?? ""}
          src={notification.actor.avatar_url}
          size="sm"
        />
      ) : (
        <BellIcon className="text-muted-foreground size-4 shrink-0" />
      )}
      {identifier && (
        <span className="text-muted-foreground w-16 shrink-0 truncate font-mono text-xs">
          {identifier}
        </span>
      )}
      <div className="flex min-w-0 flex-1 items-baseline gap-2">
        <span className={cn("truncate", isUnread && "font-medium")}>{title}</span>
        <span className="text-muted-foreground truncate text-xs">
          {notificationSummary(notification)}
        </span>
      </div>
      <span className="text-muted-foreground shrink-0 text-xs">
        {isSnoozed && snoozedUntil ? (
          <>
            Until <RelativeTime value={snoozedUntil} />
          </>
        ) : (
          <RelativeTime value={notification.created_at} />
        )}
      </span>
      <span className="hidden shrink-0 group-has-hover/row:group-hover/row:flex group-data-active/row:flex">
        <Button
          variant="ghost"
          size="icon"
          aria-label={isArchived ? "Move to inbox" : "Archive"}
          onClick={(event) => {
            event.stopPropagation();
            onToggleArchive(notification);
          }}
        >
          {isArchived ? (
            <ArchiveRestoreIcon className="size-4" />
          ) : (
            <ArchiveIcon className="size-4" />
          )}
        </Button>
      </span>
    </div>
  );
}
