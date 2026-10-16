import { createFileRoute } from "@tanstack/react-router";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/import")({
  component: ImportSettingsPage,
});

function ImportSettingsPage() {
  return (
    <SettingsPage title="Import" description="Import a Linear export into this workspace.">
      <NotBuiltYet />
    </SettingsPage>
  );
}
