import { createFileRoute } from "@tanstack/react-router";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/labels")({
  component: LabelsSettingsPage,
});

function LabelsSettingsPage() {
  return (
    <SettingsPage title="Labels" description="Labels that only this team uses.">
      <NotBuiltYet />
    </SettingsPage>
  );
}
