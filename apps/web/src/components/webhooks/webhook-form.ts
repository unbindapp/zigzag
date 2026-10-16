import { z } from "zod";

export const ALL_TEAMS = "all";

export const WEBHOOK_EVENTS = [
  "issue.created",
  "issue.updated",
  "issue.deleted",
  "comment.created",
  "project.updated",
  "cycle.updated",
] as const;

export type TWebhookEvent = (typeof WEBHOOK_EVENTS)[number];

export const webhookEventLabels: Record<TWebhookEvent, string> = {
  "issue.created": "Issue created",
  "issue.updated": "Issue updated",
  "issue.deleted": "Issue deleted",
  "comment.created": "Comment created",
  "project.updated": "Project updated",
  "cycle.updated": "Cycle updated",
};

export const webhookFormSchema = z.object({
  url: z
    .string()
    .trim()
    .min(1, "URL is required.")
    .max(2048, "Keep the URL under 2048 characters.")
    .refine((value) => /^https?:\/\/[^\s/]+/i.test(value), "Use an http or https URL."),
  label: z.string().trim().max(100, "Keep the label under 100 characters."),
  events: z.array(z.enum(WEBHOOK_EVENTS)).min(1, "Pick at least one event."),
  team: z.string(),
  enabled: z.boolean(),
});

export type TWebhookFormValues = z.infer<typeof webhookFormSchema>;

export function toCreateInput(workspaceId: string, values: TWebhookFormValues) {
  return {
    workspace_id: workspaceId,
    url: values.url.trim(),
    label: values.label.trim(),
    events: values.events,
    enabled: values.enabled,
    team_id: values.team === ALL_TEAMS ? undefined : values.team,
  };
}
