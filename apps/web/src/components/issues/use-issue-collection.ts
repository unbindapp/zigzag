import { useInfiniteQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import { resolveDisplay, type TDisplaySettings } from "@/components/issues/display";
import {
  buildListIssuesInput,
  isHiddenCompleted,
  type ListIssuesFilters,
} from "@/components/issues/filters";
import { groupIssues, type TIssueGroup } from "@/components/issues/grouping";
import type { TNewIssueDefaults } from "@/components/issues/new-issue-store";
import { useTeams } from "@/components/workspace/teams-provider";
import { useNow } from "@/components/providers/now-provider";
import { issuesListQuery } from "@/lib/queries/issues";

const HOUR_MS = 60 * 60 * 1000;

export type TIssueCollectionProps = {
  workspaceId: string;
  filters: ListIssuesFilters;
  display?: Partial<TDisplaySettings>;
  fixedFilters?: Partial<ListIssuesFilters>;
};

export function useIssueCollection({
  workspaceId,
  filters,
  display: displayOverrides,
  fixedFilters,
}: TIssueCollectionProps) {
  const display = resolveDisplay(displayOverrides);
  const input = buildListIssuesInput({ workspaceId, filters, fixedFilters, display });
  const query = useInfiniteQuery(issuesListQuery(input));
  const now = useNow();
  const { teams } = useTeams();
  const fixedTeamId = fixedFilters?.team_ids?.length === 1 ? fixedFilters.team_ids[0] : undefined;
  const fixedTeam = teams.find((team) => team.id === fixedTeamId);
  const nowMs = Math.floor(now / HOUR_MS) * HOUR_MS;

  const issues = useMemo(
    () =>
      (query.data?.pages ?? [])
        .flatMap((page) => page.data)
        .filter((issue) => !isHiddenCompleted(issue, display.show_completed, nowMs)),
    [query.data, display.show_completed, nowMs],
  );

  const fixedStateTypes = input.state_types;
  const groups = useMemo(() => {
    const states = fixedTeam?.states.filter(
      (state) => !fixedStateTypes || fixedStateTypes.includes(state.type),
    );
    return groupIssues({
      issues,
      groupBy: display.group_by,
      orderBy: display.order_by,
      order: display.order,
      states,
    });
  }, [issues, display.group_by, display.order_by, display.order, fixedTeam, fixedStateTypes]);

  return {
    query,
    issues,
    groups,
    display,
    fixedTeamId,
    total: query.data?.pages[0]?.total ?? 0,
  };
}

export function newIssueDefaultsForGroup(
  group: TIssueGroup,
  base: TNewIssueDefaults,
): TNewIssueDefaults {
  const meta = group.meta;
  switch (meta.kind) {
    case "none":
      return base;
    case "state":
      return {
        ...base,
        teamId: meta.state.team_id,
        stateId: meta.state.id,
        draft: { ...base.draft, state: meta.state },
      };
    case "priority":
      return { ...base, draft: { ...base.draft, priority: meta.priority } };
    case "assignee":
      return { ...base, draft: { ...base.draft, assignee: meta.user } };
    case "project":
      return {
        ...base,
        projectId: meta.project?.id,
        draft: { ...base.draft, project: meta.project },
      };
    case "cycle":
      return { ...base, cycleId: meta.cycle?.id, draft: { ...base.draft, cycle: meta.cycle } };
    case "label":
      return { ...base, draft: { ...base.draft, labels: meta.label ? [meta.label] : [] } };
    case "parent":
      return { ...base, parentId: meta.parent?.id, draft: { ...base.draft, parent: meta.parent } };
  }
}
