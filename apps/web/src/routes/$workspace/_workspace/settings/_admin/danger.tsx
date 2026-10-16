import { createFileRoute, useNavigate } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useDeleteWorkspaceMutation } from "@/lib/queries/workspaces";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/danger")({
  component: WorkspaceDangerSettingsPage,
});

function WorkspaceDangerSettingsPage() {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { mutateAsync: deleteWorkspace } = useDeleteWorkspaceMutation();

  return (
    <SettingsPage title="Danger zone" description="Actions here cannot be undone.">
      <SettingsSection
        variant="destructive"
        title="Delete workspace"
        description="Deletes the workspace with every team, issue, project and document in it."
        footer={
          <ConfirmDialog
            trigger={
              <Button variant="destructive" size="sm">
                Delete workspace
              </Button>
            }
            title={`Delete ${workspace.name}`}
            description="Everything in this workspace is deleted permanently for every member."
            confirmLabel="Delete workspace"
            confirmText={workspace.slug}
            onConfirm={async () => {
              await deleteWorkspace({ workspaceId: workspace.id, confirmSlug: workspace.slug });
              await navigate({ to: "/" });
            }}
          />
        }
      />
    </SettingsPage>
  );
}
