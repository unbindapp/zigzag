import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { BlocksIcon } from "lucide-react";

import { ConnectedAppRow } from "@/components/connected-apps/connected-app-row";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { CopyField } from "@/components/shared/copy-button";
import { connectedAppsListQuery } from "@/lib/queries/connected-apps";

export const Route = createFileRoute("/$workspace/_workspace/settings/account/connected-apps")({
  component: AccountConnectedAppsSettingsPage,
});

function AccountConnectedAppsSettingsPage() {
  const { data, isPending, error, refetch } = useQuery(connectedAppsListQuery());
  const apps = data?.connectedApps ?? [];
  return (
    <SettingsPage
      title="Connected apps"
      description="Applications you authorized to use Zigzag on your behalf."
    >
      <SettingsSection
        title="MCP server"
        description="Add this URL to an MCP client. It asks you to approve access here."
      >
        <CopyField value={`${window.location.origin}/mcp`} label="MCP server URL" />
      </SettingsSection>
      <SettingsSection title="Authorized apps">
        <ListStates
          isPending={isPending}
          error={error}
          onRetry={refetch}
          isEmpty={apps.length === 0}
          empty={{ icon: BlocksIcon, title: "No connected apps yet" }}
        >
          <List>
            {apps.map((app) => (
              <ConnectedAppRow key={app.id} app={app} />
            ))}
          </List>
        </ListStates>
      </SettingsSection>
    </SettingsPage>
  );
}
