import { createFileRoute } from "@tanstack/react-router";

import { LinearImport } from "@/components/import/linear-import";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/import")({
  component: ImportSettingsPage,
});

function ImportSettingsPage() {
  return (
    <SettingsPage title="Import" description="Import a Linear export into this workspace.">
      <LinearImport />
    </SettingsPage>
  );
}
