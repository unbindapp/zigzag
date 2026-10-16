import {
  ArchiveIcon,
  ArchiveRestoreIcon,
  BellIcon,
  BoxIcon,
  InboxIcon,
  MailIcon,
} from "lucide-react";

import { notificationSummary } from "@/components/inbox/notification-text";
import { SnoozeMenu } from "@/components/inbox/snooze-menu";
import { IssueDetail } from "@/components/issues/issue-detail";
import { EmptyState } from "@/components/layout/empty-state";
import { Button, LinkButton } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import type { TNotification } from "@/lib/queries/notifications";

type TProps = {
  notification: TNotification | undefined;
  isArchived: boolean;
  isSnoozeOpen: boolean;
  onSnoozeOpenChange: (open: boolean) => void;
  onSnooze: (notification: TNotification, until: Date) => void;
  onMarkUnread: (notification: TNotification) => void;
  onToggleArchive: (notification: TNotification) => void;
};

export function NotificationPane({
  notification,
  isArchived,
  isSnoozeOpen,
  onSnoozeOpenChange,
  onSnooze,
  onMarkUnread,
  onToggleArchive,
}: TProps) {
  if (!notification) {
    return (
      <section aria-label="Notification" className="flex min-h-0 w-full flex-1 flex-col">
        <EmptyState
          icon={InboxIcon}
          title="No notification selected"
          description="Pick one from the list, or move with J and K."
        />
      </section>
    );
  }
  return (
    <section
      aria-label={notification.title}
      className="bg-background flex min-h-0 w-full flex-1 flex-col"
    >
      <div className="flex h-10 shrink-0 items-center gap-2 border-b px-3">
        <span className="text-muted-foreground min-w-0 flex-1 truncate text-xs">
          {notificationSummary(notification)}
        </span>
        <Button
          variant="ghost"
          size="icon"
          aria-label="Mark unread"
          disabled={!notification.read_at}
          onClick={() => onMarkUnread(notification)}
        >
          <MailIcon className="size-4" />
        </Button>
        <SnoozeMenu
          open={isSnoozeOpen}
          onOpenChange={onSnoozeOpenChange}
          onSnooze={(until) => onSnooze(notification, until)}
        />
        <Button
          variant="ghost"
          size="icon"
          aria-label={isArchived ? "Move to inbox" : "Archive"}
          onClick={() => onToggleArchive(notification)}
        >
          {isArchived ? (
            <ArchiveRestoreIcon className="size-4" />
          ) : (
            <ArchiveIcon className="size-4" />
          )}
        </Button>
      </div>
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">
        <NotificationTarget notification={notification} />
      </div>
    </section>
  );
}

function NotificationTarget({ notification }: { notification: TNotification }) {
  const { workspace } = useWorkspace();
  if (notification.issue) {
    return (
      <IssueDetail identifier={notification.issue.identifier} variant="peek" shortcuts={false} />
    );
  }
  if (notification.project) {
    return (
      <EmptyState
        icon={BoxIcon}
        title={notification.project.name}
        description={notificationSummary(notification)}
        action={
          <LinkButton
            to="/$workspace/project/$slug"
            params={{ workspace: workspace.slug, slug: notification.project.slug }}
            variant="outline"
            size="sm"
          >
            Open project
          </LinkButton>
        }
      />
    );
  }
  return (
    <EmptyState
      icon={BellIcon}
      title={notification.title}
      description={notificationSummary(notification)}
    />
  );
}
