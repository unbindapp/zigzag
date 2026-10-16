import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { WebhookIcon } from "lucide-react";

import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { CreateWebhookDialog } from "@/components/webhooks/create-webhook-dialog";
import { WebhookRow } from "@/components/webhooks/webhook-row";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { webhooksListQuery } from "@/lib/queries/webhooks";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/webhooks")({
  component: WebhooksSettingsPage,
});

function WebhooksSettingsPage() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(webhooksListQuery(workspace.id));
  const webhooks = data?.webhooks ?? [];
  return (
    <SettingsPage title="Webhooks" description="Send events to other services.">
      <SettingsSection
        title="Webhooks"
        description="Each delivery is a signed JSON POST, retried 3 times when the endpoint fails."
        actions={<CreateWebhookDialog />}
      >
        <ListStates
          isPending={isPending}
          error={error}
          onRetry={refetch}
          isEmpty={webhooks.length === 0}
          empty={{
            icon: WebhookIcon,
            title: "No webhooks yet",
            description: "Add one to send issue, comment, project and cycle events to a URL.",
          }}
        >
          <List>
            {webhooks.map((webhook) => (
              <WebhookRow key={webhook.id} webhook={webhook} />
            ))}
          </List>
        </ListStates>
      </SettingsSection>
    </SettingsPage>
  );
}
