import { useQuery } from "@tanstack/react-query";
import { FolderGit2Icon, RefreshCwIcon, UnplugIcon } from "lucide-react";
import { useState } from "react";

import { GithubIcon } from "@/components/icons/github";
import { GithubRepositoryRow } from "@/components/integrations/github-repository-row";
import { GithubSecretDialog } from "@/components/integrations/github-secret-dialog";
import { GithubWebhookSteps } from "@/components/integrations/github-webhook-steps";
import { ErrorCard } from "@/components/layout/error-card";
import { List } from "@/components/layout/list";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { ListStates } from "@/components/layout/list-states";
import { PropertyRow } from "@/components/layout/property-row";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { CopyField } from "@/components/shared/copy-button";
import { RelativeTime } from "@/components/shared/relative-time";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  githubIntegrationQuery,
  type TGithubIntegration,
  type TGithubSecret,
  useCreateGithubIntegrationMutation,
  useDeleteGithubIntegrationMutation,
  useRegenerateGithubSecretMutation,
} from "@/lib/queries/integrations";

const description =
  "Pull requests and branches that mention an issue identifier link to the issue and move it through the workflow.";

export function GithubSettings() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(githubIntegrationQuery(workspace.id));
  const [secret, setSecret] = useState<TGithubSecret | null>(null);

  if (isPending) return <ListSkeleton />;
  if (error) return <ErrorCard message={error.message} onRetry={() => void refetch()} />;

  const integration = data.status.integration;
  return (
    <>
      {integration ? (
        <ConnectedGithub integration={integration} onSecret={setSecret} />
      ) : (
        <ConnectGithub onSecret={setSecret} />
      )}
      <GithubSecretDialog secret={secret} onClose={() => setSecret(null)} />
    </>
  );
}

function ConnectGithub({ onSecret }: { onSecret: (secret: TGithubSecret) => void }) {
  const { workspace } = useWorkspace();
  const { mutate: connect, isPending } = useCreateGithubIntegrationMutation();
  return (
    <SettingsSection
      title="GitHub"
      description={description}
      actions={
        <Button
          size="sm"
          isPending={isPending}
          onClick={() =>
            connect({ workspaceId: workspace.id }, { onSuccess: ({ data }) => onSecret(data) })
          }
        >
          <GithubIcon />
          Connect GitHub
        </Button>
      }
    />
  );
}

type TConnectedProps = {
  integration: TGithubIntegration;
  onSecret: (secret: TGithubSecret) => void;
};

function ConnectedGithub({ integration, onSecret }: TConnectedProps) {
  const { workspace } = useWorkspace();
  const { mutateAsync: regenerate } = useRegenerateGithubSecretMutation();
  const { mutateAsync: disconnect } = useDeleteGithubIntegrationMutation();
  const repositories = integration.repositories;

  return (
    <>
      <SettingsSection
        title="GitHub"
        description={description}
        actions={
          <>
            <ConfirmDialog
              trigger={
                <Button variant="outline" size="sm">
                  <RefreshCwIcon className="size-4" />
                  Regenerate
                </Button>
              }
              title="Regenerate the webhook secret"
              description="GitHub deliveries signed with the current secret are refused until you paste the new one into the webhook on GitHub."
              confirmLabel="Regenerate secret"
              variant="default"
              onConfirm={async () => {
                const { data } = await regenerate({ workspaceId: workspace.id });
                onSecret(data);
              }}
            />
            <ConfirmDialog
              trigger={
                <Button variant="outline" size="sm">
                  <UnplugIcon className="size-4" />
                  Disconnect
                </Button>
              }
              title="Disconnect GitHub"
              description="Deliveries are refused and the repository list is cleared. Links already on issues stay."
              confirmLabel="Disconnect GitHub"
              onConfirm={() => disconnect({ workspaceId: workspace.id })}
            />
          </>
        }
      >
        <CopyField value={integration.webhook_url} label="Webhook URL" />
        <PropertyRow label="Secret">
          <span className="font-mono text-xs">{integration.secret_prefix}…</span>
        </PropertyRow>
        <PropertyRow label="Last delivery">
          {integration.last_delivery_at ? (
            <RelativeTime value={integration.last_delivery_at} />
          ) : (
            <span className="text-muted-foreground">Never</span>
          )}
        </PropertyRow>
        <GithubWebhookSteps />
      </SettingsSection>
      <SettingsSection
        title="Repositories"
        description="Repositories whose webhooks delivered events with a valid signature."
      >
        <ListStates
          isPending={false}
          error={null}
          isEmpty={repositories.length === 0}
          empty={{
            icon: FolderGit2Icon,
            title: "No deliveries yet",
            description: "GitHub sends a ping as soon as you add the webhook.",
          }}
        >
          <List>
            {repositories.map((repository) => (
              <GithubRepositoryRow key={repository.id} repository={repository} />
            ))}
          </List>
        </ListStates>
      </SettingsSection>
    </>
  );
}
