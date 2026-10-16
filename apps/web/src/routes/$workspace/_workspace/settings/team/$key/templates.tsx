import { createFileRoute } from "@tanstack/react-router";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/templates")({
  component: TemplatesSettingsPage,
});

function TemplatesSettingsPage() {
  return (
    <SettingsPage title="Templates" description="Templates for this team.">
      <NotBuiltYet />
    </SettingsPage>
  );
}
