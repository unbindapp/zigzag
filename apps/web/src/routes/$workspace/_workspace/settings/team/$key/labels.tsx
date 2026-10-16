import { createFileRoute } from "@tanstack/react-router";

import { LabelManager } from "@/components/labels/label-manager";
import { SettingsPage } from "@/components/layout/settings-page";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/labels")({
  component: LabelsSettingsPage,
});

function LabelsSettingsPage() {
  const { workspace, isAdmin } = useWorkspace();
  const team = useRequiredTeamFromRoute();
  return (
    <SettingsPage title="Labels" description="Labels that only this team uses.">
      <LabelManager
        workspaceId={workspace.id}
        teamId={team.id}
        canEdit={isAdmin}
        title={`${team.name} labels`}
        description="Drag to reorder. Workspace labels are managed in workspace settings."
      />
    </SettingsPage>
  );
}
