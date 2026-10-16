import { createFileRoute } from "@tanstack/react-router";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { SettingsPage } from "@/components/layout/settings-page";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/project-statuses")({
  component: ProjectStatusesSettingsPage,
});

function ProjectStatusesSettingsPage() {
  return (
    <SettingsPage title="Project statuses" description="The statuses a project moves through.">
      <NotBuiltYet />
    </SettingsPage>
  );
}
