import {
  infiniteQueryOptions,
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem } from "@/lib/server/types";

export type TNotification = TListItem<TGoClient["notifications"]["list"]>;
export type TNotificationFilter = NonNullable<TInput<TGoClient["notifications"]["list"]>["filter"]>;

const LIST_PAGE_SIZE = 50;
const LIST_REFETCH_INTERVAL_MS = 15 * 1000;
const COUNTS_REFETCH_INTERVAL_MS = 30 * 1000;

export const queryKeyNotifications = {
  all: () => ["notifications"] as const,
  lists: () => ["notifications", "list"] as const,
  list: (workspaceId: string, filter: TNotificationFilter) =>
    ["notifications", "list", workspaceId, filter] as const,
  counts: (workspaceId: string) => ["notifications", "counts", workspaceId] as const,
};

export const notificationsListQuery = (workspaceId: string, filter: TNotificationFilter) =>
  infiniteQueryOptions({
    queryKey: queryKeyNotifications.list(workspaceId, filter),
    queryFn: async ({ pageParam }) => {
      const res = await getGoClient().notifications.list({
        workspace_id: workspaceId,
        filter,
        limit: LIST_PAGE_SIZE,
        cursor: pageParam,
      });
      return { notifications: res.data ?? [], nextCursor: res.next_cursor ?? null };
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.nextCursor ?? undefined,
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const notificationCountsQuery = (workspaceId: string) =>
  queryOptions({
    queryKey: queryKeyNotifications.counts(workspaceId),
    queryFn: async () => {
      const res = await getGoClient().notifications.counts({ workspace_id: workspaceId });
      return { unread: res.data.unread };
    },
    refetchInterval: COUNTS_REFETCH_INTERVAL_MS,
  });

type TIdsInput = { notificationIds: string[] };

export async function markNotificationsRead({ notificationIds }: TIdsInput) {
  const res = await getGoClient().notifications.read({ notification_ids: notificationIds });
  return { data: res.data };
}

export async function markNotificationsUnread({ notificationIds }: TIdsInput) {
  const res = await getGoClient().notifications.unread({ notification_ids: notificationIds });
  return { data: res.data };
}

export async function archiveNotifications({ notificationIds }: TIdsInput) {
  const res = await getGoClient().notifications.archive({ notification_ids: notificationIds });
  return { data: res.data };
}

export async function unarchiveNotifications({ notificationIds }: TIdsInput) {
  const res = await getGoClient().notifications.unarchive({ notification_ids: notificationIds });
  return { data: res.data };
}

export async function snoozeNotifications(input: TIdsInput & { until: Date }) {
  const res = await getGoClient().notifications.snooze({
    notification_ids: input.notificationIds,
    until: input.until.toISOString(),
  });
  return { data: res.data };
}

export async function markAllNotificationsRead({ workspaceId }: { workspaceId: string }) {
  const res = await getGoClient().notifications.readAll({ workspace_id: workspaceId });
  return { data: res.data };
}

function useInvalidateNotifications() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyNotifications.all() });
}

export function useMarkNotificationsReadMutation() {
  const invalidate = useInvalidateNotifications();
  return useMutation({
    mutationFn: markNotificationsRead,
    meta: { errorTitle: "Could not mark as read" },
    onSuccess: invalidate,
  });
}

export function useMarkNotificationsUnreadMutation() {
  const invalidate = useInvalidateNotifications();
  return useMutation({
    mutationFn: markNotificationsUnread,
    meta: { errorTitle: "Could not mark as unread" },
    onSuccess: invalidate,
  });
}

export function useArchiveNotificationsMutation() {
  const invalidate = useInvalidateNotifications();
  return useMutation({
    mutationFn: archiveNotifications,
    meta: { errorTitle: "Could not archive" },
    onSuccess: invalidate,
  });
}

export function useUnarchiveNotificationsMutation() {
  const invalidate = useInvalidateNotifications();
  return useMutation({
    mutationFn: unarchiveNotifications,
    meta: { errorTitle: "Could not move back to inbox" },
    onSuccess: invalidate,
  });
}

export function useSnoozeNotificationsMutation() {
  const invalidate = useInvalidateNotifications();
  return useMutation({
    mutationFn: snoozeNotifications,
    meta: { errorTitle: "Could not snooze" },
    onSuccess: invalidate,
  });
}

export function useMarkAllNotificationsReadMutation() {
  const invalidate = useInvalidateNotifications();
  return useMutation({
    mutationFn: markAllNotificationsRead,
    meta: { errorTitle: "Could not mark all as read" },
    onSuccess: invalidate,
  });
}
