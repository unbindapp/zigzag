import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { DownloadIcon, Trash2Icon } from "lucide-react";
import { useEffect } from "react";

import { GithubIcon } from "@/components/icons/github";
import {
  importSummary,
  isImporting,
  linkedIssuesLabel,
} from "@/components/integrations/github-sync";
import { ListRow, ListRowText } from "@/components/layout/list";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { ProjectIcon } from "@/components/shared/project-icon";
import { RelativeTime } from "@/components/shared/relative-time";
import { TeamIcon } from "@/components/shared/team-icon";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  GITHUB_POLL_MS,
  githubRepoSyncQuery,
  queryKeyIntegrations,
  type TGithubRepoSync,
  useBackfillGithubRepoSyncMutation,
  useRemoveGithubRepoSyncMutation,
} from "@/lib/queries/integrations";

export function GithubRepoSyncRow({ repoSync }: { repoSync: TGithubRepoSync }) {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const isRunning = isImporting(repoSync);
  const { data: live } = useQuery({
    ...githubRepoSyncQuery(repoSync.id),
    enabled: isRunning,
    refetchInterval: GITHUB_POLL_MS,
  });
  const current = (isRunning && live?.repoSync) || repoSync;
  const finished = isRunning && live !== undefined && !isImporting(live.repoSync);
  useEffect(() => {
    if (!finished) return;
    void queryClient.invalidateQueries({ queryKey: queryKeyIntegrations.repoSyncs(workspace.id) });
  }, [finished, queryClient, workspace.id]);

  const { mutateAsync: backfill } = useBackfillGithubRepoSyncMutation();
  const { mutateAsync: remove } = useRemoveGithubRepoSyncMutation();
  const summary = importSummary(current);

  return (
    <ListRow
      actions={
        <>
          <ImportStatus repoSync={current} summary={summary} />
          {current.last_error && current.backfill_status !== "failed" && (
            <SimpleTooltip content={current.last_error}>
              <Badge variant="warning">Sync error</Badge>
            </SimpleTooltip>
          )}
          {current.project && (
            <Link
              to="/$workspace/project/$slug/issues"
              params={{ workspace: workspace.slug, slug: current.project.slug }}
              className="text-muted-foreground hidden items-center gap-1.5 text-xs outline-hidden focus-visible:underline has-hover:hover:underline md:flex"
            >
              <ProjectIcon icon={current.project.icon} color={current.project.color} />
              {linkedIssuesLabel(current.linked_issues)}
            </Link>
          )}
          {current.team && (
            <SimpleTooltip content={`New issues go to ${current.team.name}`}>
              <span className="hidden sm:flex">
                <TeamIcon
                  teamKey={current.team.key}
                  icon={current.team.icon}
                  color={current.team.color}
                  size="sm"
                />
              </span>
            </SimpleTooltip>
          )}
          {current.last_synced_at && (
            <span className="text-muted-foreground hidden text-xs lg:inline">
              <RelativeTime value={current.last_synced_at} />
            </span>
          )}
          <ConfirmDialog
            trigger={
              <Button
                variant="ghost"
                size="icon"
                aria-label={`Import existing issues from ${current.full_name}`}
                disabled={isImporting(current)}
              >
                <DownloadIcon className="size-4" />
              </Button>
            }
            title={`Import issues from ${current.full_name}`}
            description={`Zigzag copies the ${current.sync_closed ? "open and closed" : "open"} issues and their comments in the background. Issues already linked are brought up to date.`}
            confirmLabel="Import issues"
            variant="default"
            onConfirm={() => backfill({ repoSyncId: current.id })}
          />
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Remove ${current.full_name}`}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title={`Remove ${current.full_name}`}
            description="Zigzag stops syncing this repository. Its issues stay, without the link to GitHub."
            confirmLabel="Remove repository"
            onConfirm={() => remove({ repoSyncId: current.id })}
          />
        </>
      }
    >
      <GithubIcon className="text-muted-foreground" />
      <a
        href={current.html_url}
        target="_blank"
        rel="noreferrer noopener"
        className="flex min-w-0 flex-1 outline-hidden focus-visible:underline has-hover:hover:underline"
      >
        <ListRowText primary={current.full_name} secondary={current.team?.name} />
      </a>
    </ListRow>
  );
}

function ImportStatus({
  repoSync,
  summary,
}: {
  repoSync: TGithubRepoSync;
  summary: string | null;
}) {
  switch (repoSync.backfill_status) {
    case "running":
      return (
        <SimpleTooltip content={summary ?? "Starting"}>
          <Badge variant="process">Importing</Badge>
        </SimpleTooltip>
      );
    case "failed":
      return (
        <SimpleTooltip content={repoSync.last_error || "The import failed"}>
          <Badge variant="destructive">Import failed</Badge>
        </SimpleTooltip>
      );
    case "completed":
      return (
        <SimpleTooltip content={summary ?? "Imported"}>
          <Badge variant="muted">Imported</Badge>
        </SimpleTooltip>
      );
    default:
      return null;
  }
}
