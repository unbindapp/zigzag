import { PlusIcon } from "lucide-react";
import { useState } from "react";

import { CheckboxList } from "@/components/shared/checkbox-list";
import { TeamIcon } from "@/components/shared/team-icon";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  ALL_TEAMS,
  toCreateInput,
  WEBHOOK_EVENTS,
  webhookEventLabels,
  webhookFormSchema,
  type TWebhookEvent,
  type TWebhookFormValues,
} from "@/components/webhooks/webhook-form";
import { WebhookSecretDialog } from "@/components/webhooks/webhook-secret-dialog";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm, useFieldError } from "@/lib/hooks/use-app-form";
import { type TWebhookCreated, useCreateWebhookMutation } from "@/lib/queries/webhooks";

export function CreateWebhookDialog() {
  const [isOpen, setIsOpen] = useState(false);
  const [created, setCreated] = useState<TWebhookCreated | null>(null);
  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger
          render={
            <Button size="sm">
              <PlusIcon className="size-4" />
              Add webhook
            </Button>
          }
        />
        <DialogContent className="w-lg max-w-full">
          <DialogHeader>
            <DialogTitle>Add webhook</DialogTitle>
            <DialogDescription>
              Zigzag sends a signed JSON POST to the URL for every event you pick.
            </DialogDescription>
          </DialogHeader>
          {isOpen && (
            <CreateWebhookForm
              onCreated={(webhook) => {
                setIsOpen(false);
                setCreated(webhook);
              }}
            />
          )}
        </DialogContent>
      </Dialog>
      <WebhookSecretDialog created={created} onClose={() => setCreated(null)} />
    </>
  );
}

function CreateWebhookForm({ onCreated }: { onCreated: (webhook: TWebhookCreated) => void }) {
  const { workspace } = useWorkspace();
  const { teams } = useTeams();
  const { mutateAsync: create } = useCreateWebhookMutation();
  const form = useAppForm({
    defaultValues: {
      url: "",
      label: "",
      events: ["issue.created", "issue.updated"] as TWebhookEvent[],
      team: ALL_TEAMS,
      enabled: true,
    } satisfies TWebhookFormValues as TWebhookFormValues,
    validators: { onSubmit: webhookFormSchema },
    onSubmit: async ({ value }) => {
      const res = await create(toCreateInput(workspace.id, value));
      onCreated(res.data);
    },
  });

  const teamItems = [
    { value: ALL_TEAMS, label: "All teams" },
    ...teams.map((team) => ({
      value: team.id,
      label: team.name,
      icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
    })),
  ];

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="url"
        children={(field) => (
          <field.TextField
            label="URL"
            description="A public http or https endpoint. Private and local addresses are refused."
            placeholder="https://example.com/hooks/zigzag"
            autoFocus
          />
        )}
      />
      <form.AppField
        name="label"
        children={(field) => <field.TextField label="Label" placeholder="Slack bridge" />}
      />
      <form.Field name="events" children={(field) => <EventsField field={field} />} />
      <form.AppField
        name="team"
        children={(field) => (
          <field.SelectField
            label="Team"
            description="Send events of one team only, or of the whole workspace."
            items={teamItems}
          />
        )}
      />
      <form.AppField
        name="enabled"
        children={(field) => (
          <field.SwitchField label="Enabled" description="Turn off to pause deliveries." />
        )}
      />
      <DialogFooter>
        <DialogClose
          render={
            <Button type="button" variant="outline" size="sm">
              Cancel
            </Button>
          }
        />
        <form.AppForm>
          <form.SubmitButton size="sm">Add webhook</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}

type TEventsFieldApi = Parameters<typeof useFieldError>[0];

function EventsField({ field }: { field: TEventsFieldApi }) {
  const { error } = useFieldError(field);
  return (
    <div className="flex flex-col gap-1.5">
      <p className="text-sm font-medium">Events</p>
      <CheckboxList
        label="Events"
        items={WEBHOOK_EVENTS.map((event) => ({
          value: event,
          label: webhookEventLabels[event],
          description: event,
        }))}
        value={field.state.value as string[]}
        onChange={(value) => {
          field.handleChange(value);
          field.handleBlur();
        }}
      />
      {error && (
        <p role="alert" className="text-destructive text-xs">
          {error}
        </p>
      )}
    </div>
  );
}
