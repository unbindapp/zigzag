import { useQuery } from "@tanstack/react-query";
import { FolderSyncIcon } from "lucide-react";

import { AddGithubRepoSyncDialog } from "@/components/integrations/add-github-repo-sync-dialog";
import { GithubRepoSyncRow } from "@/components/integrations/github-repo-sync-row";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsSection } from "@/components/layout/settings-section";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { githubRepoSyncsQuery } from "@/lib/queries/integrations";

export function GithubRepoSyncs() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(githubRepoSyncsQuery(workspace.id));
  const repoSyncs = data?.repoSyncs ?? [];
  return (
    <SettingsSection
      title="Synced repositories"
      description="Issues, comments, labels, assignees and state sync both ways with these repositories."
      actions={<AddGithubRepoSyncDialog />}
    >
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={() => void refetch()}
        isEmpty={repoSyncs.length === 0}
        empty={{
          icon: FolderSyncIcon,
          title: "No synced repositories yet",
          description: "Add a repository to bring its GitHub issues into a team.",
        }}
      >
        <List>
          {repoSyncs.map((repoSync) => (
            <GithubRepoSyncRow key={repoSync.id} repoSync={repoSync} />
          ))}
        </List>
      </ListStates>
    </SettingsSection>
  );
}
