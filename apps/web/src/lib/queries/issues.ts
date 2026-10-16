import {
  type InfiniteData,
  infiniteQueryOptions,
  type QueryClient,
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";

import type { TIssuePatch as TIssueDraftPatch } from "@/components/issues/issue-draft";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

type TIssuesApi = TGoClient["issues"];

export type TIssue = TListItem<TIssuesApi["list"]>;
export type TIssueDetail = TResponseData<TIssuesApi["get"]>;
export type TIssuePage = Awaited<ReturnType<TIssuesApi["list"]>>;
export type TListIssuesInput = TInput<TIssuesApi["list"]>;
export type TCreateIssueInput = TInput<TIssuesApi["create"]>;
export type TUpdateIssueInput = TInput<TIssuesApi["update"]>;
export type TBulkUpdateIssuesInput = TInput<TIssuesApi["bulkUpdate"]>;
export type TComment = TListItem<TIssuesApi["comments"]["list"]>;
export type TCommentReply = TComment["replies"][number];
export type THistoryEntry = TListItem<TIssuesApi["history"]["list"]>;
export type TFieldChange = THistoryEntry["changes"][number];
export type TAttachment = TIssueDetail["attachments"][number];
export type TReactionGroup = TIssueDetail["reactions"][number];
export type TRelatedIssue = TIssueDetail["relations"]["blocks"][number];
export type TUserSummary = NonNullable<TIssue["assignee"]>;
export type TWorkflowStateSummary = TIssue["state"];
export type TLabelSummary = TIssue["labels"][number];
export type TOrderBy = NonNullable<TListIssuesInput["order_by"]>;
export type TGroupBy = NonNullable<TListIssuesInput["group_by"]>;

export const ISSUE_TITLE_MAX_LENGTH = 255;
export const CLEAR_UUID = "00000000-0000-0000-0000-000000000000";
export const CLEAR_DATE = "";
export const CLEAR_ESTIMATE = -1;

const LIST_PAGE_SIZE = 100;
const SEARCH_LIMIT = 8;
const HISTORY_PAGE_SIZE = 100;
export const LIST_REFETCH_INTERVAL_MS = 15 * 1000;

function issuesApi() {
  return getGoClient().issues;
}

export const queryKeyIssues = {
  all: () => ["issues"] as const,
  lists: () => ["issues", "list"] as const,
  list: (input: TListIssuesInput) => ["issues", "list", input] as const,
  details: () => ["issues", "detail"] as const,
  detail: (workspaceId: string, identifier: string) =>
    ["issues", "detail", workspaceId, identifier.toUpperCase()] as const,
  history: (issueId: string) => ["issues", "history", issueId] as const,
  search: (workspaceId: string, query: string) => ["issues", "search", workspaceId, query] as const,
};

const queryKeyNotificationCounts = () => ["notifications", "counts"] as const;
const queryKeyProjects = () => ["projects"] as const;
const queryKeyCycles = () => ["cycles"] as const;

export const issuesListQuery = (input: TListIssuesInput) =>
  infiniteQueryOptions({
    queryKey: queryKeyIssues.list(input),
    queryFn: async ({ pageParam }) => {
      const res = await issuesApi().list({ ...input, limit: LIST_PAGE_SIZE, cursor: pageParam });
      return { data: res.data ?? [], next_cursor: res.next_cursor ?? null, total: res.total ?? 0 };
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last: TIssuePage) => last.next_cursor ?? undefined,
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const issueQuery = (workspaceId: string, identifier: string) =>
  queryOptions({
    queryKey: queryKeyIssues.detail(workspaceId, identifier),
    queryFn: async () => {
      const res = await issuesApi().get({
        workspace_id: workspaceId,
        identifier: identifier.toUpperCase(),
      });
      return { issue: res.data };
    },
  });

export const issueHistoryQuery = (issueId: string) =>
  infiniteQueryOptions({
    queryKey: queryKeyIssues.history(issueId),
    queryFn: async ({ pageParam }) => {
      const res = await issuesApi().history.list({
        issue_id: issueId,
        limit: HISTORY_PAGE_SIZE,
        cursor: pageParam,
      });
      return { data: res.data ?? [], next_cursor: res.next_cursor ?? null };
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.next_cursor ?? undefined,
  });

export const searchIssuesQuery = (workspaceId: string, query: string) =>
  queryOptions({
    queryKey: queryKeyIssues.search(workspaceId, query),
    queryFn: async () => {
      const res = await issuesApi().search({
        workspace_id: workspaceId,
        query,
        limit: SEARCH_LIMIT,
      });
      return { issues: res.data ?? [] };
    },
    enabled: query.trim().length > 0,
  });

export async function createIssue(input: TCreateIssueInput) {
  const res = await issuesApi().create(input);
  return { data: res.data };
}

export function issuePatchToUpdateInput(
  issueId: string,
  patch: TIssueDraftPatch,
): TUpdateIssueInput {
  const input: TUpdateIssueInput = { issue_id: issueId };
  if (patch.state) input.state_id = patch.state.id;
  if (patch.priority !== undefined) input.priority = patch.priority;
  if (patch.assignee !== undefined) input.assignee_id = patch.assignee?.id ?? CLEAR_UUID;
  if (patch.labels) input.label_ids = patch.labels.map((label) => label.id);
  if (patch.project !== undefined) input.project_id = patch.project?.id ?? CLEAR_UUID;
  if (patch.milestone !== undefined) input.milestone_id = patch.milestone?.id ?? CLEAR_UUID;
  if (patch.cycle !== undefined) input.cycle_id = patch.cycle?.id ?? CLEAR_UUID;
  if (patch.estimate !== undefined) input.estimate = patch.estimate ?? CLEAR_ESTIMATE;
  if (patch.due_date !== undefined) input.due_date = patch.due_date ?? CLEAR_DATE;
  if (patch.parent !== undefined) input.parent_id = patch.parent?.id ?? CLEAR_UUID;
  return input;
}

export async function updateIssue(input: TUpdateIssueInput) {
  const res = await issuesApi().update(input);
  return { data: res.data };
}

export async function bulkUpdateIssues(input: TBulkUpdateIssuesInput) {
  const res = await issuesApi().bulkUpdate(input);
  return { data: res.data };
}

export async function deleteIssue(input: { issueId: string }) {
  const res = await issuesApi().delete({ issue_id: input.issueId });
  return { data: res.data };
}

type TIssueLifecycleAction = "archive" | "unarchive" | "trash" | "restore";

export async function changeIssueLifecycle(input: {
  issueIds: string[];
  action: TIssueLifecycleAction;
}) {
  const api = issuesApi();
  const results = await Promise.all(
    input.issueIds.map((issueId) => api[input.action]({ issue_id: issueId })),
  );
  return { data: results.map((res) => res.data) };
}

export async function setIssueSubscription(input: {
  issueId: string;
  subscribed: boolean;
  userId?: string;
}) {
  const api = issuesApi();
  const call = input.subscribed ? api.subscribe : api.unsubscribe;
  const res = await call({ issue_id: input.issueId, user_id: input.userId });
  return { data: res.data };
}

export type TReactionTarget = { issueId: string } | { commentId: string };

export async function toggleReaction(input: {
  emoji: string;
  target: TReactionTarget;
  reacted: boolean;
}) {
  const api = issuesApi();
  const call = input.reacted ? api.reactions.remove : api.reactions.add;
  const target =
    "issueId" in input.target
      ? { issue_id: input.target.issueId }
      : { comment_id: input.target.commentId };
  const res = await call({ emoji: input.emoji, ...target });
  return { data: res.data };
}

export async function createAttachment(input: {
  issueId: string;
  title: string;
  url?: string;
  uploadId?: string;
}) {
  const res = await issuesApi().attachments.create({
    issue_id: input.issueId,
    title: input.title,
    url: input.url,
    upload_id: input.uploadId,
  });
  return { data: res.data };
}

export async function deleteAttachment(input: { attachmentId: string }) {
  const res = await issuesApi().attachments.delete({ attachment_id: input.attachmentId });
  return { data: res.data };
}

export function invalidateIssues(queryClient: QueryClient) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: queryKeyIssues.all() }),
    queryClient.invalidateQueries({ queryKey: queryKeyNotificationCounts() }),
    queryClient.invalidateQueries({ queryKey: queryKeyProjects() }),
    queryClient.invalidateQueries({ queryKey: queryKeyCycles() }),
  ]);
}

function useInvalidateIssues() {
  const queryClient = useQueryClient();
  return () => invalidateIssues(queryClient);
}

type TIssuePatch = { id: string; patch: Partial<TIssue> };

function patchIssueCaches(queryClient: QueryClient, patches: TIssuePatch[]) {
  const byId = new Map(patches.map((item) => [item.id, item.patch]));
  const apply = <T extends TIssue>(issue: T): T => {
    const patch = byId.get(issue.id);
    return patch ? { ...issue, ...patch } : issue;
  };
  queryClient.setQueriesData<InfiniteData<TIssuePage>>(
    { queryKey: queryKeyIssues.lists() },
    (data) =>
      data && {
        ...data,
        pages: data.pages.map((page) => ({ ...page, data: page.data.map(apply) })),
      },
  );
  queryClient.setQueriesData<{ issue: TIssueDetail }>(
    { queryKey: queryKeyIssues.details() },
    (data) =>
      data && { issue: { ...apply(data.issue), children: data.issue.children?.map(apply) } },
  );
}

type TIssueCacheSnapshot = [readonly unknown[], unknown][];

async function optimisticallyPatch(
  queryClient: QueryClient,
  patches: TIssuePatch[],
): Promise<TIssueCacheSnapshot> {
  if (patches.length === 0) return [];
  await queryClient.cancelQueries({ queryKey: queryKeyIssues.all() });
  const snapshot: TIssueCacheSnapshot = [
    ...queryClient.getQueriesData({ queryKey: queryKeyIssues.lists() }),
    ...queryClient.getQueriesData({ queryKey: queryKeyIssues.details() }),
  ];
  patchIssueCaches(queryClient, patches);
  return snapshot;
}

function restoreIssueCaches(queryClient: QueryClient, snapshot: TIssueCacheSnapshot | undefined) {
  snapshot?.forEach(([queryKey, data]) => queryClient.setQueryData(queryKey, data));
}

export type TUpdateIssueVariables = { input: TUpdateIssueInput; optimistic?: Partial<TIssue> };

export function useCreateIssueMutation() {
  const invalidate = useInvalidateIssues();
  return useMutation({
    mutationFn: createIssue,
    meta: { errorTitle: "Could not create issue" },
    onSuccess: invalidate,
  });
}

export function useUpdateIssueMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (variables: TUpdateIssueVariables) => updateIssue(variables.input),
    meta: { errorTitle: "Could not update issue" },
    onMutate: (variables) =>
      optimisticallyPatch(
        queryClient,
        variables.optimistic ? [{ id: variables.input.issue_id, patch: variables.optimistic }] : [],
      ),
    onError: (_error, _variables, snapshot) => restoreIssueCaches(queryClient, snapshot),
    onSettled: () => invalidateIssues(queryClient),
  });
}

export type TBulkUpdateIssuesVariables = {
  input: TBulkUpdateIssuesInput;
  optimistic?: Partial<TIssue>;
};

export function useBulkUpdateIssuesMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (variables: TBulkUpdateIssuesVariables) => bulkUpdateIssues(variables.input),
    meta: { errorTitle: "Could not update issues" },
    onMutate: (variables) => {
      const patch = variables.optimistic;
      return optimisticallyPatch(
        queryClient,
        patch ? (variables.input.issue_ids ?? []).map((id) => ({ id, patch })) : [],
      );
    },
    onError: (_error, _variables, snapshot) => restoreIssueCaches(queryClient, snapshot),
    onSettled: () => invalidateIssues(queryClient),
  });
}

export function useDeleteIssueMutation() {
  const invalidate = useInvalidateIssues();
  return useMutation({
    mutationFn: deleteIssue,
    meta: { errorTitle: "Could not delete issue" },
    onSuccess: invalidate,
  });
}

export function useIssueLifecycleMutation() {
  const invalidate = useInvalidateIssues();
  return useMutation({
    mutationFn: changeIssueLifecycle,
    meta: { errorTitle: "Could not update issue" },
    onSuccess: invalidate,
  });
}

export function useIssueSubscriptionMutation() {
  const invalidate = useInvalidateIssues();
  return useMutation({
    mutationFn: setIssueSubscription,
    meta: { errorTitle: "Could not change subscription" },
    onSuccess: invalidate,
  });
}

export function useToggleReactionMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: toggleReaction,
    meta: { errorTitle: "Could not update reaction" },
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeyIssues.details() }),
        queryClient.invalidateQueries({ queryKey: queryKeyNotificationCounts() }),
      ]),
  });
}

export function useCreateAttachmentMutation() {
  const invalidate = useInvalidateIssues();
  return useMutation({
    mutationFn: createAttachment,
    meta: { errorTitle: "Could not add attachment" },
    onSuccess: invalidate,
  });
}

export function useDeleteAttachmentMutation() {
  const invalidate = useInvalidateIssues();
  return useMutation({
    mutationFn: deleteAttachment,
    meta: { errorTitle: "Could not remove attachment" },
    onSuccess: invalidate,
  });
}
