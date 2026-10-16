import { createFileRoute, Outlet } from "@tanstack/react-router";
import {
  ArchiveIcon,
  FileStackIcon,
  LockIcon,
  RefreshCwIcon,
  SettingsIcon,
  TagIcon,
  TriangleAlertIcon,
  TriangleIcon,
  UsersIcon,
  WorkflowIcon,
  RulerIcon,
} from "lucide-react";

import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { PageContainer } from "@/components/layout/page-container";
import { SettingsLayout, type TSettingsTab } from "@/components/layout/settings-layout";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { EntityNotFound } from "@/components/navigation/not-found-template";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key")({
  component: TeamSettingsLayout,
});

function TeamSettingsLayout() {
  const { key } = Route.useParams();
  const { workspace, isAdmin } = useWorkspace();
  const { teamByKey, isPending, error, refetch } = useTeams();
  const team = teamByKey(key);

  if (!isAdmin) {
    return (
      <PageContainer width="narrow">
        <EmptyState
          icon={LockIcon}
          title="Admins only"
          description="Ask a workspace admin to change team settings."
        />
      </PageContainer>
    );
  }
  if (isPending) {
    return (
      <PageContainer width="narrow">
        <ListSkeleton />
      </PageContainer>
    );
  }
  if (error) {
    return (
      <PageContainer width="narrow">
        <ErrorCard message={error.message} onRetry={refetch} />
      </PageContainer>
    );
  }
  if (!team) {
    return (
      <PageContainer width="narrow">
        <EntityNotFound
          Icon={UsersIcon}
          description="There is no team with this key in this workspace."
        />
      </PageContainer>
    );
  }

  const params = { workspace: workspace.slug, key: team.key };
  const tabs: TSettingsTab[] = [
    {
      label: "General",
      icon: SettingsIcon,
      link: { to: "/$workspace/settings/team/$key", params },
      exact: true,
    },
    {
      label: "Workflow",
      icon: WorkflowIcon,
      link: { to: "/$workspace/settings/team/$key/workflow", params },
    },
    {
      label: "Cycles",
      icon: RefreshCwIcon,
      link: { to: "/$workspace/settings/team/$key/cycles", params },
    },
    {
      label: "Estimates",
      icon: RulerIcon,
      link: { to: "/$workspace/settings/team/$key/estimates", params },
    },
    {
      label: "Triage",
      icon: TriangleIcon,
      link: { to: "/$workspace/settings/team/$key/triage", params },
    },
    {
      label: "Members",
      icon: UsersIcon,
      link: { to: "/$workspace/settings/team/$key/members", params },
    },
    {
      label: "Labels",
      icon: TagIcon,
      link: { to: "/$workspace/settings/team/$key/labels", params },
    },
    {
      label: "Templates",
      icon: FileStackIcon,
      link: { to: "/$workspace/settings/team/$key/templates", params },
    },
    {
      label: "Auto close & archive",
      icon: ArchiveIcon,
      link: { to: "/$workspace/settings/team/$key/auto-archive", params },
    },
    {
      label: "Danger zone",
      icon: TriangleAlertIcon,
      link: { to: "/$workspace/settings/team/$key/danger", params },
    },
  ];
  return (
    <SettingsLayout title={team.name} tabs={tabs}>
      <Outlet />
    </SettingsLayout>
  );
}
