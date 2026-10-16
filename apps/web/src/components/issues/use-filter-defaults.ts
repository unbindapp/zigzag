import type { ListIssuesFilters } from "@/components/issues/filters";
import { UNASSIGNED } from "@/components/issues/filters";
import type { TIssueDraft } from "@/components/issues/issue-draft";
import type { TNewIssueDefaults } from "@/components/issues/new-issue-store";
import {
  useCycles,
  useLabels,
  useProjects,
  useTeamStates,
  useWorkspaceUsers,
} from "@/components/issues/picker-sources";
import { flattenLabels } from "@/components/issues/pickers";

function single<T>(values: T[] | undefined) {
  return values?.length === 1 ? values[0] : undefined;
}

export function useFilterDefaults(
  filters: ListIssuesFilters,
  teamId: string | undefined,
): TNewIssueDefaults {
  const { users } = useWorkspaceUsers();
  const states = useTeamStates(teamId);
  const { labels } = useLabels(teamId);
  const { projects } = useProjects(teamId);
  const { cycles } = useCycles(teamId);

  const draft: Partial<TIssueDraft> = {};
  const stateId = single(filters.state_ids);
  const stateType = single(filters.state_types);
  const state =
    states.find((item) => item.id === stateId) ??
    (stateType ? states.find((item) => item.type === stateType) : undefined);
  if (state) draft.state = state;
  const priority = single(filters.priorities);
  if (priority !== undefined) draft.priority = priority;
  const assigneeId = single(filters.assignee_ids);
  const assignee =
    assigneeId && assigneeId !== UNASSIGNED ? users.find((u) => u.id === assigneeId) : undefined;
  if (assignee) draft.assignee = assignee;
  const labelIds = filters.label_ids ?? [];
  const flat = flattenLabels(labels).map((item) => item.label);
  const filterLabels = flat.filter((label) => labelIds.includes(label.id));
  if (filterLabels.length > 0) {
    draft.labels = filterLabels.map((label) => ({
      id: label.id,
      name: label.name,
      color: label.color,
      parent_id: label.parent_id,
    }));
  }
  const project = projects.find((item) => item.id === single(filters.project_ids));
  if (project) {
    draft.project = {
      id: project.id,
      name: project.name,
      slug: project.slug,
      icon: project.icon,
      color: project.color,
    };
  }
  const cycle = cycles.find((item) => item.id === single(filters.cycle_ids));
  if (cycle) draft.cycle = { id: cycle.id, number: cycle.number, name: cycle.name };

  return {
    teamId,
    stateId: draft.state?.id,
    projectId: draft.project?.id,
    cycleId: draft.cycle?.id,
    draft,
  };
}
