export type TNotificationType =
  | "issue_assigned"
  | "issue_unassigned"
  | "issue_mentioned"
  | "comment_mentioned"
  | "issue_commented"
  | "issue_state_changed"
  | "issue_priority_changed"
  | "issue_due_soon"
  | "issue_blocking"
  | "project_update"
  | "project_added_as_member"
  | "project_lead"
  | "reaction";

export const notificationTypeLabels: Record<TNotificationType, string> = {
  issue_assigned: "Assigned to you",
  issue_unassigned: "Unassigned from you",
  issue_mentioned: "Mentioned you",
  comment_mentioned: "Mentioned you in a comment",
  issue_commented: "New comment",
  issue_state_changed: "Status changed",
  issue_priority_changed: "Priority changed",
  issue_due_soon: "Due soon",
  issue_blocking: "Blocking issue changed",
  project_update: "Project update",
  project_added_as_member: "Added you to a project",
  project_lead: "Made you project lead",
  reaction: "Reacted",
};

type TNotificationLike = {
  type: TNotificationType;
  title: string;
  body: string;
  actor?: { name: string; display_name: string };
  issue?: { identifier: string };
};

export function notificationTitle(notification: Pick<TNotificationLike, "title" | "issue">) {
  const identifier = notification.issue?.identifier;
  if (!identifier || !notification.title.startsWith(`${identifier} `)) return notification.title;
  return notification.title.slice(identifier.length + 1);
}

export function notificationSummary(notification: TNotificationLike) {
  const body = notification.body.trim();
  if (body) return body;
  const label = notificationTypeLabels[notification.type];
  const actor = notification.actor?.name || notification.actor?.display_name;
  return actor ? `${actor}: ${label.toLowerCase()}` : label;
}
