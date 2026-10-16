import { Link } from "@tanstack/react-router";
import { ChevronRightIcon, PlusIcon } from "lucide-react";

import { formatProjectDate } from "@/components/projects/project-dates";
import { AvatarStack } from "@/components/projects/project-pickers";
import { ProgressBar } from "@/components/projects/progress-bar";
import { ProjectStatusIcon } from "@/components/projects/project-status-icon";
import { HealthIcon, healthLabels } from "@/components/projects/health-badge";
import { Avatar } from "@/components/shared/avatar";
import { PriorityIcon } from "@/components/shared/priority-icon";
import { ProjectIcon } from "@/components/shared/project-icon";
import { Button } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import type { TProjectListItem, TProjectStatus } from "@/lib/queries/projects";

type TRowProps = {
  workspaceSlug: string;
  project: TProjectListItem;
};

export function ProjectRow({ workspaceSlug, project }: TRowProps) {
  const lead = project.lead;
  return (
    <div
      role="listitem"
      className="has-hover:hover:bg-accent flex h-9 w-full items-center gap-2 border-b px-3 text-sm last:border-b-0"
    >
      <ProjectIcon icon={project.icon} color={project.color} />
      <Link
        to="/$workspace/project/$slug"
        params={{ workspace: workspaceSlug, slug: project.slug }}
        className="min-w-0 flex-1 truncate font-medium"
      >
        {project.name}
      </Link>
      {project.health && (
        <SimpleTooltip content={healthLabels[project.health]}>
          <span className="hidden shrink-0 items-center sm:flex">
            <HealthIcon health={project.health} />
          </span>
        </SimpleTooltip>
      )}
      <span className="text-muted-foreground hidden w-32 shrink-0 items-center gap-1.5 truncate text-xs md:flex">
        <ProjectStatusIcon type={project.status.type} color={project.status.color} />
        <span className="truncate">{project.status.name}</span>
      </span>
      <span className="flex w-6 shrink-0 justify-center">
        {lead && (
          <SimpleTooltip content={`Lead: ${lead.name || lead.display_name}`}>
            <span className="flex">
              <Avatar
                id={lead.id}
                name={lead.name || lead.display_name}
                src={lead.avatar_url}
                size="sm"
              />
            </span>
          </SimpleTooltip>
        )}
      </span>
      <span className="hidden w-16 shrink-0 lg:flex">
        {project.members.length > 0 && <AvatarStack users={project.members} />}
      </span>
      <span className="text-muted-foreground hidden w-24 shrink-0 truncate text-xs sm:block">
        {project.target_date
          ? formatProjectDate(project.target_date, project.target_date_resolution)
          : ""}
      </span>
      <span className="w-28 shrink-0">
        <ProgressBar percent={project.progress.percent} label={`${project.name} progress`} />
      </span>
      <span className="flex w-6 shrink-0 justify-center">
        <PriorityIcon priority={project.priority} />
      </span>
    </div>
  );
}

type THeaderProps = {
  status: TProjectStatus;
  count: number;
  isCollapsed: boolean;
  onToggle: () => void;
  onCreate?: () => void;
};

export function ProjectGroupHeader({
  status,
  count,
  isCollapsed,
  onToggle,
  onCreate,
}: THeaderProps) {
  return (
    <div
      role="listitem"
      aria-label={`${status.name}, ${count} projects`}
      className="bg-muted flex h-9 w-full items-center gap-2 border-b px-1.5 text-sm"
    >
      <Button
        variant="ghost"
        size="sm"
        forceMinSize={false}
        aria-expanded={!isCollapsed}
        onClick={onToggle}
        className="min-w-0 justify-start"
      >
        <ChevronRightIcon
          data-open={!isCollapsed || undefined}
          className="text-muted-foreground size-3.5 transition-transform data-open:rotate-90"
        />
        <ProjectStatusIcon type={status.type} color={status.color} />
        <span className="truncate">{status.name}</span>
        <span className="text-muted-foreground text-xs font-normal">{count}</span>
      </Button>
      <span className="flex-1" />
      {onCreate && (
        <Button
          variant="ghost"
          size="icon"
          aria-label={`New project in ${status.name}`}
          onClick={onCreate}
        >
          <PlusIcon className="size-4" />
        </Button>
      )}
    </div>
  );
}
