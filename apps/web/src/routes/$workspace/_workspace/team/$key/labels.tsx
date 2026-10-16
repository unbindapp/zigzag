import { createFileRoute } from "@tanstack/react-router";
import { TagIcon } from "lucide-react";

import { LabelManager } from "@/components/labels/label-manager";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { teamCrumb } from "@/components/workspace/team-crumb";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/labels")({
  component: TeamKeyLabelsPage,
});

function TeamKeyLabelsPage() {
  const { workspace, isAdmin } = useWorkspace();
  const team = useRequiredTeamFromRoute();
  return (
    <PageContainer>
      <PageHeader
        title="Labels"
        icon={<TagIcon />}
        breadcrumb={[teamCrumb(team, workspace.slug), { label: "Labels" }]}
      />
      <LabelManager
        workspaceId={workspace.id}
        teamId={team.id}
        canEdit={isAdmin}
        title={`${team.name} labels`}
        description="Labels only this team uses."
      />
      <LabelManager
        workspaceId={workspace.id}
        canEdit={isAdmin}
        title="Workspace labels"
        description="Labels every team shares."
      />
    </PageContainer>
  );
}
