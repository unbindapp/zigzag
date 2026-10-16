import type {
  TCreateIssueInput,
  TIssue,
  TLabelSummary,
  TUserSummary,
  TWorkflowStateSummary,
} from "@/lib/queries/issues";

export type TIssueDraft = {
  state: TWorkflowStateSummary | undefined;
  priority: number;
  assignee: TUserSummary | null;
  labels: TLabelSummary[];
  project: NonNullable<TIssue["project"]> | null;
  milestone: NonNullable<TIssue["milestone"]> | null;
  cycle: NonNullable<TIssue["cycle"]> | null;
  estimate: number | null;
  due_date: string | null;
  parent: NonNullable<TIssue["parent"]> | null;
};

export type TIssuePatch = Partial<TIssueDraft>;

export const EMPTY_DRAFT: TIssueDraft = {
  state: undefined,
  priority: 0,
  assignee: null,
  labels: [],
  project: null,
  milestone: null,
  cycle: null,
  estimate: null,
  due_date: null,
  parent: null,
};

export function draftFromIssue(issue: TIssue): TIssueDraft {
  return {
    state: issue.state,
    priority: issue.priority,
    assignee: issue.assignee ?? null,
    labels: issue.labels,
    project: issue.project ?? null,
    milestone: issue.milestone ?? null,
    cycle: issue.cycle ?? null,
    estimate: issue.estimate,
    due_date: issue.due_date,
    parent: issue.parent ?? null,
  };
}

export function applyPatch(draft: TIssueDraft, patch: TIssuePatch): TIssueDraft {
  const next = { ...draft, ...patch };
  if ("project" in patch && patch.project?.id !== draft.project?.id) next.milestone = null;
  return next;
}

export function patchToOptimistic(patch: TIssuePatch): Partial<TIssue> {
  const optimistic: Partial<TIssue> = {};
  if (patch.state) optimistic.state = patch.state;
  if (patch.priority !== undefined) optimistic.priority = patch.priority;
  if (patch.assignee !== undefined) optimistic.assignee = patch.assignee ?? undefined;
  if (patch.labels) optimistic.labels = patch.labels;
  if (patch.project !== undefined) {
    optimistic.project = patch.project ?? undefined;
    optimistic.milestone = undefined;
  }
  if (patch.milestone !== undefined) optimistic.milestone = patch.milestone ?? undefined;
  if (patch.cycle !== undefined) optimistic.cycle = patch.cycle ?? undefined;
  if (patch.estimate !== undefined) optimistic.estimate = patch.estimate;
  if (patch.due_date !== undefined) optimistic.due_date = patch.due_date;
  if (patch.parent !== undefined) optimistic.parent = patch.parent ?? undefined;
  return optimistic;
}

export function draftToCreateInput({
  teamId,
  title,
  description,
  draft,
}: {
  teamId: string;
  title: string;
  description: string;
  draft: TIssueDraft;
}): TCreateIssueInput {
  const input: TCreateIssueInput = { team_id: teamId, title: title.trim() };
  if (description.trim()) input.description = description;
  if (draft.state) input.state_id = draft.state.id;
  if (draft.priority) input.priority = draft.priority;
  if (draft.assignee) input.assignee_id = draft.assignee.id;
  if (draft.labels.length > 0) input.label_ids = draft.labels.map((label) => label.id);
  if (draft.project) input.project_id = draft.project.id;
  if (draft.milestone) input.milestone_id = draft.milestone.id;
  if (draft.cycle) input.cycle_id = draft.cycle.id;
  if (draft.estimate !== null) input.estimate = draft.estimate;
  if (draft.due_date) input.due_date = draft.due_date;
  if (draft.parent) input.parent_id = draft.parent.id;
  return input;
}
