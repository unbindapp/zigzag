import { useHotkey } from "@tanstack/react-hotkeys";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { ArchiveIcon, CheckCheckIcon, InboxIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { NotificationPane } from "@/components/inbox/notification-pane";
import { NotificationRow } from "@/components/inbox/notification-row";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { SplitView } from "@/components/layout/split-view";
import { useDeviceSize } from "@/components/providers/device-size-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTab } from "@/components/ui/tabs";
import { toast } from "@/components/ui/toast";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  notificationCountsQuery,
  notificationsListQuery,
  type TNotification,
  useArchiveNotificationsMutation,
  useMarkAllNotificationsReadMutation,
  useMarkNotificationsReadMutation,
  useMarkNotificationsUnreadMutation,
  useSnoozeNotificationsMutation,
  useUnarchiveNotificationsMutation,
} from "@/lib/queries/notifications";
import { formatDateTime } from "@/lib/helpers/dates";

export const INBOX_TABS = ["inbox", "read", "archived", "snoozed"] as const;
export type TInboxTab = (typeof INBOX_TABS)[number];

const tabLabels: Record<TInboxTab, string> = {
  inbox: "Inbox",
  read: "Read",
  archived: "Archived",
  snoozed: "Snoozed",
};

const emptyStates: Record<TInboxTab, { title: string; description: string }> = {
  inbox: {
    title: "You are all caught up",
    description: "Assignments, mentions and comments on issues you follow show up here.",
  },
  read: { title: "Nothing read yet", description: "Notifications you opened stay here." },
  archived: { title: "Nothing archived", description: "Archive a notification with E." },
  snoozed: { title: "Nothing snoozed", description: "Snooze a notification with H." },
};

type TProps = {
  tab: TInboxTab;
  selectedId: string | undefined;
  onTabChange: (tab: TInboxTab) => void;
  onSelect: (id: string | undefined) => void;
};

export function InboxView({ tab, selectedId, onTabChange, onSelect }: TProps) {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { size } = useDeviceSize();
  const hasPane = size === "lg" || size === "xl";
  const [isSnoozeOpen, setIsSnoozeOpen] = useState(false);

  const list = useInfiniteQuery(notificationsListQuery(workspace.id, tab));
  const { data: counts } = useQuery(notificationCountsQuery(workspace.id));
  const notifications = list.data?.pages.flatMap((page) => page.notifications) ?? [];
  const selectedIndex = notifications.findIndex((item) => item.id === selectedId);
  const selected = selectedIndex >= 0 ? notifications[selectedIndex] : undefined;
  const isArchivedTab = tab === "archived";
  const unread = counts?.unread ?? 0;

  const { mutate: markRead } = useMarkNotificationsReadMutation();
  const { mutate: markUnread } = useMarkNotificationsUnreadMutation();
  const { mutate: archive } = useArchiveNotificationsMutation();
  const { mutate: unarchive } = useUnarchiveNotificationsMutation();
  const { mutate: snooze } = useSnoozeNotificationsMutation();
  const { mutate: markAllRead, isPending: isMarkingAll } = useMarkAllNotificationsReadMutation();

  const readOnOpen = useRef<string | null>(null);
  useEffect(() => {
    if (!selected || selected.read_at || readOnOpen.current === selected.id) return;
    readOnOpen.current = selected.id;
    markRead({ notificationIds: [selected.id] });
  }, [selected, markRead]);

  const neighbourOf = (notification: TNotification) => {
    const index = notifications.findIndex((item) => item.id === notification.id);
    return notifications[index + 1] ?? notifications[index - 1];
  };

  const open = (notification: TNotification) => {
    if (hasPane) {
      onSelect(notification.id);
      return;
    }
    if (!notification.read_at) markRead({ notificationIds: [notification.id] });
    if (notification.issue) {
      void navigate({
        to: "/$workspace/issue/$identifier",
        params: { workspace: workspace.slug, identifier: notification.issue.identifier },
      });
      return;
    }
    if (notification.project) {
      void navigate({
        to: "/$workspace/project/$slug",
        params: { workspace: workspace.slug, slug: notification.project.slug },
      });
    }
  };

  const leave = (notification: TNotification) => {
    if (notification.id !== selectedId) return;
    onSelect(neighbourOf(notification)?.id);
  };

  const toggleArchive = (notification: TNotification) => {
    leave(notification);
    const input = { notificationIds: [notification.id] };
    if (isArchivedTab) {
      unarchive(input, {
        onSuccess: () => toast.add({ type: "success", title: "Moved to inbox" }),
      });
      return;
    }
    archive(input, {
      onSuccess: () =>
        toast.add({
          type: "success",
          title: "Archived",
          data: {
            action: (
              <Button variant="outline" size="sm" onClick={() => unarchive(input)}>
                Undo
              </Button>
            ),
          },
        }),
    });
  };

  const snoozeUntil = (notification: TNotification, until: Date) => {
    leave(notification);
    snooze(
      { notificationIds: [notification.id], until },
      {
        onSuccess: () =>
          toast.add({
            type: "success",
            title: "Snoozed",
            description: `Until ${formatDateTime(until)}`,
          }),
      },
    );
  };

  const move = (step: number) => {
    if (notifications.length === 0) return;
    const next = selectedIndex < 0 ? 0 : selectedIndex + step;
    const target = notifications[Math.min(Math.max(next, 0), notifications.length - 1)];
    onSelect(target.id);
    if (next >= notifications.length - 1 && list.hasNextPage && !list.isFetchingNextPage) {
      void list.fetchNextPage();
    }
  };

  const shortcutsEnabled = !isSnoozeOpen;
  useHotkey("J", () => move(1), { enabled: shortcutsEnabled });
  useHotkey("K", () => move(-1), { enabled: shortcutsEnabled });
  useHotkey("E", () => selected && toggleArchive(selected), {
    enabled: shortcutsEnabled && Boolean(selected),
  });
  useHotkey("U", () => selected && markUnread({ notificationIds: [selected.id] }), {
    enabled: shortcutsEnabled && Boolean(selected?.read_at),
  });
  useHotkey("H", () => setIsSnoozeOpen(true), { enabled: shortcutsEnabled && Boolean(selected) });

  const main = (
    <PageContainer>
      <PageHeader
        title="Inbox"
        icon={<InboxIcon />}
        actions={
          <Button
            variant="outline"
            disabled={unread === 0}
            isPending={isMarkingAll}
            onClick={() => markAllRead({ workspaceId: workspace.id })}
          >
            <CheckCheckIcon className="size-4" />
            Mark all read
          </Button>
        }
        toolbar={
          <Tabs value={tab} onValueChange={(value) => onTabChange(value as TInboxTab)}>
            <TabsList>
              {INBOX_TABS.map((value) => (
                <TabsTab key={value} value={value}>
                  {tabLabels[value]}
                  {value === "inbox" && unread > 0 && (
                    <Badge variant="muted">{unread > 99 ? "99+" : unread}</Badge>
                  )}
                </TabsTab>
              ))}
            </TabsList>
          </Tabs>
        }
      />
      <ListStates
        isPending={list.isPending}
        error={list.error}
        onRetry={() => void list.refetch()}
        isEmpty={notifications.length === 0}
        empty={{ icon: tab === "archived" ? ArchiveIcon : InboxIcon, ...emptyStates[tab] }}
      >
        <div className="flex w-full flex-col gap-2">
          <List>
            {notifications.map((notification) => (
              <NotificationRow
                key={notification.id}
                notification={notification}
                isActive={notification.id === selectedId}
                isArchived={isArchivedTab}
                onOpen={open}
                onToggleArchive={toggleArchive}
              />
            ))}
          </List>
          {list.hasNextPage && (
            <Button
              variant="ghost"
              size="sm"
              className="self-center"
              isPending={list.isFetchingNextPage}
              onClick={() => void list.fetchNextPage()}
            >
              Load more
            </Button>
          )}
        </div>
      </ListStates>
    </PageContainer>
  );

  return (
    <SplitView
      storageKey="inbox-pane-width"
      main={main}
      aside={
        hasPane ? (
          <NotificationPane
            notification={selected}
            isArchived={isArchivedTab}
            isSnoozeOpen={isSnoozeOpen}
            onSnoozeOpenChange={setIsSnoozeOpen}
            onSnooze={snoozeUntil}
            onMarkUnread={(notification) => {
              readOnOpen.current = notification.id;
              markUnread({ notificationIds: [notification.id] });
            }}
            onToggleArchive={toggleArchive}
          />
        ) : undefined
      }
    />
  );
}
