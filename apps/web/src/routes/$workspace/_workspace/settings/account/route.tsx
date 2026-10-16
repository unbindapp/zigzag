import { createFileRoute, Outlet } from "@tanstack/react-router";
import { BellIcon, BlocksIcon, KeyRoundIcon, KeySquareIcon, UserIcon } from "lucide-react";

import { SettingsLayout, type TSettingsTab } from "@/components/layout/settings-layout";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/settings/account")({
  component: AccountSettingsLayout,
});

function AccountSettingsLayout() {
  const { workspace } = useWorkspace();
  const params = { workspace: workspace.slug };
  const tabs: TSettingsTab[] = [
    {
      label: "Profile",
      icon: UserIcon,
      link: { to: "/$workspace/settings/account", params },
      exact: true,
    },
    {
      label: "Password",
      icon: KeyRoundIcon,
      link: { to: "/$workspace/settings/account/password", params },
    },
    {
      label: "Notifications",
      icon: BellIcon,
      link: { to: "/$workspace/settings/account/notifications", params },
    },
    {
      label: "API keys",
      icon: KeySquareIcon,
      link: { to: "/$workspace/settings/account/api-keys", params },
    },
    {
      label: "Connected apps",
      icon: BlocksIcon,
      link: { to: "/$workspace/settings/account/connected-apps", params },
    },
  ];
  return (
    <SettingsLayout title="Account" tabs={tabs}>
      <Outlet />
    </SettingsLayout>
  );
}
