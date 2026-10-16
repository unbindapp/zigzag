import { createFileRoute, Outlet } from "@tanstack/react-router";
import { LockIcon, MailIcon, ServerIcon, UsersIcon } from "lucide-react";

import { EmptyState } from "@/components/layout/empty-state";
import { PageContainer } from "@/components/layout/page-container";
import { SettingsLayout, type TSettingsTab } from "@/components/layout/settings-layout";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/settings/system")({
  component: SystemSettingsLayout,
});

function SystemSettingsLayout() {
  const { workspace, isSystemAdmin } = useWorkspace();
  if (!isSystemAdmin) {
    return (
      <PageContainer width="narrow">
        <EmptyState
          icon={LockIcon}
          title="System admins only"
          description="Instance settings are managed by system admins."
        />
      </PageContainer>
    );
  }
  const params = { workspace: workspace.slug };
  const tabs: TSettingsTab[] = [
    {
      label: "General",
      icon: ServerIcon,
      link: { to: "/$workspace/settings/system", params },
      exact: true,
    },
    { label: "Email", icon: MailIcon, link: { to: "/$workspace/settings/system/email", params } },
    { label: "Users", icon: UsersIcon, link: { to: "/$workspace/settings/system/users", params } },
  ];
  return (
    <SettingsLayout title="Instance" tabs={tabs}>
      <Outlet />
    </SettingsLayout>
  );
}
