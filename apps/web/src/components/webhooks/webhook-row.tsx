import { SendIcon, Trash2Icon, WebhookIcon } from "lucide-react";

import { ListRow, ListRowText } from "@/components/layout/list";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { TeamIcon } from "@/components/shared/team-icon";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { toast } from "@/components/ui/toast";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useTeams } from "@/components/workspace/teams-provider";
import {
  type TWebhook,
  useDeleteWebhookMutation,
  useTestWebhookMutation,
  useUpdateWebhookMutation,
} from "@/lib/queries/webhooks";

export function WebhookRow({ webhook }: { webhook: TWebhook }) {
  const { teams } = useTeams();
  const team = teams.find((item) => item.id === webhook.team_id);
  const name = webhook.label || webhook.url;
  const { mutate: update, isPending: isUpdating } = useUpdateWebhookMutation();
  const { mutateAsync: remove } = useDeleteWebhookMutation();
  const { mutate: test, isPending: isTesting } = useTestWebhookMutation();

  const sendTest = () =>
    test(
      { webhookId: webhook.id },
      {
        onSuccess: ({ data }) =>
          toast.add(
            data.delivered
              ? {
                  type: "success",
                  title: "Test event delivered",
                  description: `HTTP ${data.status_code}`,
                }
              : { type: "error", title: "Test event failed", description: data.error },
          ),
      },
    );

  return (
    <ListRow
      muted={!webhook.enabled}
      actions={
        <>
          {team && (
            <SimpleTooltip content={`Only ${team.name}`}>
              <span className="hidden sm:flex">
                <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />
              </span>
            </SimpleTooltip>
          )}
          <SimpleTooltip content={webhook.events.join(", ")}>
            <Badge variant="muted" className="hidden sm:inline-flex">
              {webhook.events.length === 1 ? "1 event" : `${webhook.events.length} events`}
            </Badge>
          </SimpleTooltip>
          <Switch
            aria-label={`${webhook.enabled ? "Disable" : "Enable"} ${name}`}
            checked={webhook.enabled}
            disabled={isUpdating}
            onCheckedChange={(enabled) => update({ webhook_id: webhook.id, enabled })}
          />
          <Button
            variant="ghost"
            size="icon"
            aria-label={`Send a test event to ${name}`}
            isPending={isTesting}
            onClick={sendTest}
          >
            <SendIcon className="size-4" />
          </Button>
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Delete ${name}`}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title={`Delete ${name}`}
            description="No more events are sent to this URL."
            confirmLabel="Delete"
            onConfirm={() => remove({ webhookId: webhook.id })}
          />
        </>
      }
    >
      <WebhookIcon className="text-muted-foreground size-4 shrink-0" />
      <ListRowText primary={name} secondary={webhook.label ? webhook.url : undefined} />
    </ListRow>
  );
}
