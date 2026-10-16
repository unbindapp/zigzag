import { createFileRoute } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { TemplateManager } from "@/components/templates/template-manager";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/templates")({
  component: TemplatesSettingsPage,
});

function TemplatesSettingsPage() {
  const { workspace, isAdmin } = useWorkspace();
  return (
    <SettingsPage title="Templates" description="Templates for new issues, projects and documents.">
      <TemplateManager
        workspaceId={workspace.id}
        canEdit={isAdmin}
        title="Issue templates"
        description="Every team can start an issue from these."
      />
    </SettingsPage>
  );
}
