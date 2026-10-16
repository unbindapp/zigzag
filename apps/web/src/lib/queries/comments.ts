import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { invalidateIssues, toggleReaction } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";

function commentsApi() {
  return getGoClient().issues.comments;
}

const COMMENTS_LIMIT = 250;

export const queryKeyComments = {
  all: () => ["comments"] as const,
  list: (issueId: string) => ["comments", "list", issueId] as const,
};

export const commentsQuery = (issueId: string) =>
  queryOptions({
    queryKey: queryKeyComments.list(issueId),
    queryFn: async () => {
      const res = await commentsApi().list({ issue_id: issueId, limit: COMMENTS_LIMIT });
      return { comments: res.data ?? [] };
    },
  });

export async function createComment(input: { issueId: string; body: string; parentId?: string }) {
  const res = await commentsApi().create({
    issue_id: input.issueId,
    body: input.body,
    parent_id: input.parentId,
  });
  return { data: res.data };
}

export async function updateComment(input: { commentId: string; body: string }) {
  const res = await commentsApi().update({ comment_id: input.commentId, body: input.body });
  return { data: res.data };
}

export async function deleteComment(input: { commentId: string }) {
  const res = await commentsApi().delete({ comment_id: input.commentId });
  return { data: res.data };
}

export async function setCommentResolved(input: { commentId: string; resolved: boolean }) {
  const api = commentsApi();
  const call = input.resolved ? api.resolve : api.unresolve;
  const res = await call({ comment_id: input.commentId });
  return { data: res.data };
}

function useInvalidateComments() {
  const queryClient = useQueryClient();
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeyComments.all() }),
      invalidateIssues(queryClient),
    ]);
}

export function useCreateCommentMutation() {
  const invalidate = useInvalidateComments();
  return useMutation({
    mutationFn: createComment,
    meta: { errorTitle: "Could not post comment" },
    onSuccess: invalidate,
  });
}

export function useUpdateCommentMutation() {
  const invalidate = useInvalidateComments();
  return useMutation({
    mutationFn: updateComment,
    meta: { errorTitle: "Could not edit comment" },
    onSuccess: invalidate,
  });
}

export function useDeleteCommentMutation() {
  const invalidate = useInvalidateComments();
  return useMutation({
    mutationFn: deleteComment,
    meta: { errorTitle: "Could not delete comment" },
    onSuccess: invalidate,
  });
}

export function useResolveCommentMutation() {
  const invalidate = useInvalidateComments();
  return useMutation({
    mutationFn: setCommentResolved,
    meta: { errorTitle: "Could not update comment" },
    onSuccess: invalidate,
  });
}

export function useToggleCommentReactionMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { commentId: string; emoji: string; reacted: boolean }) =>
      toggleReaction({
        emoji: input.emoji,
        reacted: input.reacted,
        target: { commentId: input.commentId },
      }),
    meta: { errorTitle: "Could not update reaction" },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeyComments.all() }),
  });
}
