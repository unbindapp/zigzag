import { createFileRoute } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { TemplateManager } from "@/components/templates/template-manager";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/templates")({
  component: TemplatesSettingsPage,
});

function TemplatesSettingsPage() {
  const { workspace, isAdmin } = useWorkspace();
  const team = useRequiredTeamFromRoute();
  return (
    <SettingsPage title="Templates" description="Templates for this team.">
      <TemplateManager
        workspaceId={workspace.id}
        teamId={team.id}
        canEdit={isAdmin}
        title={`${team.name} issue templates`}
        description="Team templates can also preset the status of new issues."
      />
    </SettingsPage>
  );
}
