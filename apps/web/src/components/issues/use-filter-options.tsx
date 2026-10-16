import { CircleUserIcon } from "lucide-react";

import { UNASSIGNED, type TFilterField } from "@/components/issues/filters";
import { PRIORITY_ORDER, sortStates } from "@/components/issues/grouping";
import {
  useCycles,
  useLabels,
  useProjects,
  useWorkspaceUsers,
} from "@/components/issues/picker-sources";
import { cycleName, flattenLabels, UserAvatar } from "@/components/issues/pickers";
import type { TPickerOption } from "@/components/issues/property-picker";
import { CycleIcon } from "@/components/shared/cycle-icon";
import { PriorityIcon, priorityLabel } from "@/components/shared/priority-icon";
import { ProjectIcon } from "@/components/shared/project-icon";
import { StateIcon } from "@/components/shared/state-icon";
import { useTeams } from "@/components/workspace/teams-provider";

export function useFilterOptions(teamId: string | undefined) {
  const { teams, myTeams } = useTeams();
  const { users } = useWorkspaceUsers();
  const { labels } = useLabels(teamId);
  const { projects } = useProjects(teamId);
  const { cycles, enabled: cyclesEnabled } = useCycles(teamId);

  const scopedTeams = teamId ? teams.filter((team) => team.id === teamId) : myTeams;
  const stateOptions: TPickerOption[] = scopedTeams.flatMap((team) =>
    sortStates(team.states).map((state) => ({
      value: state.id,
      label: state.name,
      group: scopedTeams.length > 1 ? team.name : undefined,
      keywords: [team.key],
      icon: <StateIcon type={state.type} color={state.color} name={state.name} />,
    })),
  );
  const userOptions: TPickerOption[] = users.map((user) => ({
    value: user.id,
    label: user.name || user.display_name,
    keywords: [user.display_name],
    icon: <UserAvatar user={user} size="xs" />,
  }));

  const options: Record<TFilterField, TPickerOption[]> = {
    status: stateOptions,
    assignee: [
      {
        value: UNASSIGNED,
        label: "No assignee",
        icon: <CircleUserIcon className="text-muted-foreground size-4" />,
      },
      ...userOptions,
    ],
    creator: userOptions,
    subscriber: userOptions,
    priority: PRIORITY_ORDER.map((priority) => ({
      value: String(priority),
      label: priorityLabel(priority),
      icon: <PriorityIcon priority={priority} />,
    })),
    labels: flattenLabels(labels).map(({ label, group }) => ({
      value: label.id,
      label: label.name,
      group: group?.name,
      icon: (
        <span
          aria-hidden
          className="size-2 shrink-0 rounded-full"
          style={{ backgroundColor: label.color }}
        />
      ),
    })),
    project: projects.map((project) => ({
      value: project.id,
      label: project.name,
      icon: <ProjectIcon icon={project.icon} color={project.color} />,
    })),
    cycle: cycles.map((cycle) => ({
      value: cycle.id,
      label: cycleName(cycle),
      icon: <CycleIcon status={cycle.is_active ? "active" : "upcoming"} />,
    })),
    due_date: [],
  };
  return { options, cyclesEnabled };
}
