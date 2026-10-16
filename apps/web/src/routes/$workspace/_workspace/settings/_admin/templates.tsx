import { createFileRoute } from "@tanstack/react-router";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/templates")({
  component: TemplatesSettingsPage,
});

function TemplatesSettingsPage() {
  return (
    <SettingsPage title="Templates" description="Templates for new issues, projects and documents.">
      <NotBuiltYet />
    </SettingsPage>
  );
}
