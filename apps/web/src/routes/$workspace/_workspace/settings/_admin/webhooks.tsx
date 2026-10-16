import { createFileRoute } from "@tanstack/react-router";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/webhooks")({
  component: WebhooksSettingsPage,
});

function WebhooksSettingsPage() {
  return (
    <SettingsPage title="Webhooks" description="Send events to other services.">
      <NotBuiltYet />
    </SettingsPage>
  );
}
