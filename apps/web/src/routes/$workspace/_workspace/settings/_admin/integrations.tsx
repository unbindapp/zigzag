import { createFileRoute } from "@tanstack/react-router";

import { GithubSettings } from "@/components/integrations/github-settings";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/integrations")({
  component: IntegrationsSettingsPage,
});

function IntegrationsSettingsPage() {
  return (
    <SettingsPage title="Integrations" description="Connect Zigzag to the tools your team uses.">
      <GithubSettings />
    </SettingsPage>
  );
}
