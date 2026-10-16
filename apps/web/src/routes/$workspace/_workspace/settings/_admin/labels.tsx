import { createFileRoute } from "@tanstack/react-router";

import { LabelManager } from "@/components/labels/label-manager";
import { SettingsPage } from "@/components/layout/settings-page";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/labels")({
  component: LabelsSettingsPage,
});

function LabelsSettingsPage() {
  const { workspace, isAdmin } = useWorkspace();
  return (
    <SettingsPage title="Labels" description="Workspace labels are shared by every team.">
      <LabelManager
        workspaceId={workspace.id}
        canEdit={isAdmin}
        title="Workspace labels"
        description="Drag to reorder. Groups collapse; a single-select group allows one of its labels per issue."
      />
    </SettingsPage>
  );
}
