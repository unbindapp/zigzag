import { useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import {
  ArchiveIcon,
  ArchiveRestoreIcon,
  BellIcon,
  BellOffIcon,
  GitBranchIcon,
  LinkIcon,
  Maximize2Icon,
  Trash2Icon,
  UndoIcon,
  XIcon,
} from "lucide-react";

import { type TIssueDetail, useIssueSubscriptionMutation } from "@/lib/queries/issues";
import {
  useCopyIssue,
  useIssueLifecycle,
  useOpenIssuePage,
} from "@/components/issues/use-issue-actions";
import { Breadcrumb } from "@/components/layout/breadcrumb";
import { TeamIcon } from "@/components/shared/team-icon";
import { Button } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { meQuery } from "@/lib/queries/me";

type TProps = {
  issue: TIssueDetail;
  variant: "page" | "peek";
  onClose?: () => void;
};

function HeaderAction({
  label,
  onClick,
  isPending,
  children,
}: {
  label: string;
  onClick: () => void;
  isPending?: boolean;
  children: ReactNode;
}) {
  return (
    <SimpleTooltip content={label}>
      <Button
        variant="ghost"
        size="icon"
        aria-label={label}
        isPending={isPending}
        onClick={onClick}
      >
        {children}
      </Button>
    </SimpleTooltip>
  );
}

export function IssueDetailHeader({ issue, variant, onClose }: TProps) {
  const { workspace, canEdit } = useWorkspace();
  const { teams } = useTeams();
  const { data: me } = useQuery(meQuery);
  const team = teams.find((item) => item.id === issue.team_id);
  const { copyLink, copyBranch } = useCopyIssue();
  const lifecycle = useIssueLifecycle();
  const openPage = useOpenIssuePage();
  const { mutate: setSubscribed, isPending: isSubscribing } = useIssueSubscriptionMutation();
  const isSubscribed = Boolean(me && issue.subscribers?.some((user) => user.id === me.user.id));

  return (
    <div className="flex w-full min-w-0 items-center gap-2">
      <div className="min-w-0 flex-1">
        <Breadcrumb
          items={[
            {
              label: team?.name ?? issue.team_key,
              icon: team ? (
                <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />
              ) : undefined,
              link: {
                to: "/$workspace/team/$key/issues",
                params: { workspace: workspace.slug, key: issue.team_key },
              },
            },
            { label: issue.identifier },
          ]}
        />
      </div>
      <div className="flex shrink-0 items-center gap-1">
        <HeaderAction label="Copy link" onClick={() => copyLink(issue)}>
          <LinkIcon className="size-4" />
        </HeaderAction>
        <HeaderAction label="Copy branch name" onClick={() => copyBranch(issue)}>
          <GitBranchIcon className="size-4" />
        </HeaderAction>
        {canEdit && (
          <HeaderAction
            label={isSubscribed ? "Unsubscribe" : "Subscribe"}
            isPending={isSubscribing}
            onClick={() => setSubscribed({ issueId: issue.id, subscribed: !isSubscribed })}
          >
            {isSubscribed ? <BellOffIcon className="size-4" /> : <BellIcon className="size-4" />}
          </HeaderAction>
        )}
        {canEdit && (
          <HeaderAction
            label={issue.archived_at ? "Unarchive" : "Archive"}
            isPending={lifecycle.isPending}
            onClick={() =>
              void lifecycle
                .run([issue.id], issue.archived_at ? "unarchive" : "archive")
                .catch(() => undefined)
            }
          >
            {issue.archived_at ? (
              <ArchiveRestoreIcon className="size-4" />
            ) : (
              <ArchiveIcon className="size-4" />
            )}
          </HeaderAction>
        )}
        {canEdit && (
          <HeaderAction
            label={issue.trashed_at ? "Restore from trash" : "Move to trash"}
            isPending={lifecycle.isPending}
            onClick={() =>
              void lifecycle
                .run([issue.id], issue.trashed_at ? "restore" : "trash")
                .catch(() => undefined)
            }
          >
            {issue.trashed_at ? <UndoIcon className="size-4" /> : <Trash2Icon className="size-4" />}
          </HeaderAction>
        )}
        {variant === "peek" && (
          <HeaderAction label="Open full page" onClick={() => void openPage(issue.identifier)}>
            <Maximize2Icon className="size-4" />
          </HeaderAction>
        )}
        {onClose && (
          <HeaderAction label="Close" onClick={onClose}>
            <XIcon className="size-4" />
          </HeaderAction>
        )}
      </div>
    </div>
  );
}
