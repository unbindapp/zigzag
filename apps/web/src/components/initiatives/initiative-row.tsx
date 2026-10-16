import { Link } from "@tanstack/react-router";
import { BoxIcon, TargetIcon } from "lucide-react";

import { formatTargetDate, userName } from "@/components/initiatives/initiative-pickers";
import { InitiativeStatusBadge } from "@/components/initiatives/initiative-status";
import { ListRow } from "@/components/layout/list";
import { HealthBadge } from "@/components/projects/health-badge";
import { ProgressBar } from "@/components/projects/progress-bar";
import { Avatar } from "@/components/shared/avatar";
import { EntityIcon } from "@/components/shared/entity-icon";
import { SimpleTooltip } from "@/components/ui/tooltip";
import type { TInitiativeListItem } from "@/lib/queries/initiatives";

type TProps = {
  workspaceSlug: string;
  initiative: TInitiativeListItem;
};

export function InitiativeRow({ workspaceSlug, initiative }: TProps) {
  const owner = initiative.owner;
  const projectCount = initiative.projects.length;
  return (
    <ListRow muted={Boolean(initiative.archived_at)}>
      <EntityIcon icon={initiative.icon} color={initiative.color} fallback={TargetIcon} />
      <Link
        to="/$workspace/initiative/$slug"
        params={{ workspace: workspaceSlug, slug: initiative.slug }}
        className="min-w-0 flex-1 truncate font-medium"
      >
        {initiative.name}
      </Link>
      <span className="hidden shrink-0 md:flex">
        <HealthBadge health={initiative.health} />
      </span>
      <span className="hidden shrink-0 sm:flex">
        <InitiativeStatusBadge status={initiative.status} />
      </span>
      <span className="text-muted-foreground hidden w-12 shrink-0 items-center gap-1.5 text-xs md:flex">
        <BoxIcon className="size-3.5" aria-hidden />
        <span aria-label={`${projectCount} projects`}>{projectCount}</span>
      </span>
      <span className="text-muted-foreground hidden w-24 shrink-0 truncate text-xs sm:block">
        {initiative.target_date ? formatTargetDate(initiative.target_date) : ""}
      </span>
      <span className="hidden w-28 shrink-0 sm:block">
        <ProgressBar percent={initiative.progress.percent} label={`${initiative.name} progress`} />
      </span>
      <span className="flex w-6 shrink-0 justify-center">
        {owner && (
          <SimpleTooltip content={`Owner: ${userName(owner)}`}>
            <span className="flex">
              <Avatar id={owner.id} name={userName(owner)} src={owner.avatar_url} size="sm" />
            </span>
          </SimpleTooltip>
        )}
      </span>
    </ListRow>
  );
}
