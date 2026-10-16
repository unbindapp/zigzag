import { createFileRoute, useNavigate } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useDeleteTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/danger")({
  component: TeamDangerSettingsPage,
});

function TeamDangerSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { mutateAsync: deleteTeam } = useDeleteTeamMutation();
  return (
    <SettingsPage title="Danger zone" description="Actions here cannot be undone.">
      <SettingsSection
        variant="destructive"
        title="Delete team"
        description="Deletes the team with all of its issues, workflow states and cycles."
        footer={
          <ConfirmDialog
            trigger={
              <Button variant="destructive" size="sm">
                Delete team
              </Button>
            }
            title={`Delete ${team.name}`}
            description="Every issue in this team is deleted permanently."
            confirmLabel="Delete team"
            confirmText={team.key}
            onConfirm={async () => {
              await deleteTeam({ teamId: team.id });
              await navigate({
                to: "/$workspace/settings/teams",
                params: { workspace: workspace.slug },
              });
            }}
          />
        }
      />
    </SettingsPage>
  );
}
