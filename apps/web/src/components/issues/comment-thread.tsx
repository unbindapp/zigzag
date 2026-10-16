import { useQuery } from "@tanstack/react-query";
import {
  CheckCircle2Icon,
  ChevronRightIcon,
  EllipsisIcon,
  PencilIcon,
  ReplyIcon,
  Trash2Icon,
  UndoIcon,
} from "lucide-react";
import { useState } from "react";

import { UserAvatar } from "@/components/issues/pickers";
import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { ReactionBar } from "@/components/issues/reaction-bar";
import type { TComment, TCommentReply } from "@/lib/queries/issues";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { MarkdownEditor } from "@/components/shared/markdown-editor";
import { MarkdownView } from "@/components/shared/markdown-view";
import { RelativeTime } from "@/components/shared/relative-time";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  useCreateCommentMutation,
  useDeleteCommentMutation,
  useResolveCommentMutation,
  useToggleCommentReactionMutation,
  useUpdateCommentMutation,
} from "@/lib/queries/comments";
import { meQuery } from "@/lib/queries/me";
import { useCreateUploadMutation } from "@/lib/queries/uploads";

export function useUploadImage() {
  const { workspace } = useWorkspace();
  const { mutateAsync } = useCreateUploadMutation();
  return async (file: File) => {
    const res = await mutateAsync({ workspaceId: workspace.id, file });
    return res.data.url;
  };
}

export function CommentComposer({
  issueId,
  parentId,
  placeholder = "Leave a comment…",
  autoFocus,
  onDone,
}: {
  issueId: string;
  parentId?: string;
  placeholder?: string;
  autoFocus?: boolean;
  onDone?: () => void;
}) {
  const [body, setBody] = useState("");
  const { users } = useWorkspaceUsers();
  const uploadImage = useUploadImage();
  const { mutateAsync: createComment, isPending } = useCreateCommentMutation();
  const submit = async () => {
    if (!body.trim() || isPending) return;
    await createComment({ issueId, body, parentId }).catch(() => undefined);
    setBody("");
    onDone?.();
  };
  return (
    <div className="bg-card flex w-full flex-col gap-2 rounded-lg border p-4">
      <MarkdownEditor
        label={parentId ? "Reply" : "Comment"}
        value={body}
        onChange={setBody}
        onSubmit={() => void submit()}
        placeholder={placeholder}
        users={users}
        onUploadImage={uploadImage}
        autoFocus={autoFocus}
      />
      <div className="flex items-center justify-end gap-2">
        {onDone && (
          <Button variant="outline" size="sm" onClick={onDone}>
            Cancel
          </Button>
        )}
        <Button
          size="sm"
          isPending={isPending}
          disabled={!body.trim()}
          onClick={() => void submit()}
        >
          {parentId ? "Reply" : "Comment"}
        </Button>
      </div>
    </div>
  );
}

export function CommentThread({ comment, issueId }: { comment: TComment; issueId: string }) {
  const [isReplying, setIsReplying] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const { canEdit } = useWorkspace();
  const isResolved = Boolean(comment.resolved_at);

  if (isResolved && !isExpanded) {
    return (
      <Button
        variant="ghost"
        size="sm"
        className="text-muted-foreground self-start"
        aria-expanded={false}
        onClick={() => setIsExpanded(true)}
      >
        <ChevronRightIcon className="size-3.5" />
        <CheckCircle2Icon className="size-4" />
        Resolved thread from {comment.author?.name ?? "a deleted user"}
        {comment.replies.length > 0 &&
          ` and ${comment.replies.length} repl${comment.replies.length === 1 ? "y" : "ies"}`}
      </Button>
    );
  }

  return (
    <div className="bg-card flex w-full flex-col rounded-lg border">
      <CommentItem
        comment={comment}
        isThreadRoot
        onReply={canEdit ? () => setIsReplying(true) : undefined}
      />
      {comment.replies.map((reply) => (
        <div key={reply.id} className="border-t">
          <CommentItem comment={reply} />
        </div>
      ))}
      {isReplying && (
        <div className="border-t p-2">
          <CommentComposer
            issueId={issueId}
            parentId={comment.id}
            placeholder="Leave a reply…"
            autoFocus
            onDone={() => setIsReplying(false)}
          />
        </div>
      )}
      {isResolved && (
        <div className="border-t px-4 py-2">
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground"
            onClick={() => setIsExpanded(false)}
          >
            Collapse resolved thread
          </Button>
        </div>
      )}
    </div>
  );
}

function CommentItem({
  comment,
  isThreadRoot,
  onReply,
}: {
  comment: TComment | TCommentReply;
  isThreadRoot?: boolean;
  onReply?: () => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [draft, setDraft] = useState(comment.body);
  const { canEdit, isAdmin } = useWorkspace();
  const { data: me } = useQuery(meQuery);
  const { users } = useWorkspaceUsers();
  const uploadImage = useUploadImage();
  const { mutateAsync: updateComment, isPending: isSaving } = useUpdateCommentMutation();
  const { mutateAsync: deleteComment } = useDeleteCommentMutation();
  const { mutate: setResolved } = useResolveCommentMutation();
  const { mutate: toggleReaction } = useToggleCommentReactionMutation();
  const isAuthor = Boolean(me && comment.author?.id === me.user.id);
  const canModify = isAuthor || isAdmin;
  const authorName = comment.author
    ? comment.author.name || comment.author.display_name
    : "Deleted user";

  const save = async () => {
    if (!draft.trim()) return;
    await updateComment({ commentId: comment.id, body: draft });
    setIsEditing(false);
  };

  return (
    <article aria-label={`Comment from ${authorName}`} className="flex flex-col gap-2 p-4">
      <header className="flex items-center gap-2">
        {comment.author && <UserAvatar user={comment.author} size="sm" />}
        <span className="truncate text-sm font-medium">{authorName}</span>
        <span className="text-muted-foreground text-xs">
          <RelativeTime value={comment.created_at} />
        </span>
        {comment.edited_at && <span className="text-muted-foreground text-xs">(edited)</span>}
        {comment.resolved_at && (
          <span className="text-muted-foreground inline-flex items-center gap-1.5 text-xs">
            <CheckCircle2Icon className="size-3.5" />
            Resolved
          </span>
        )}
        <span className="flex-1" />
        {onReply && (
          <Button variant="ghost" size="icon" aria-label="Reply" onClick={onReply}>
            <ReplyIcon className="size-4" />
          </Button>
        )}
        {(canModify || (canEdit && isThreadRoot)) && (
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button variant="ghost" size="icon" aria-label="Comment actions">
                  <EllipsisIcon className="size-4" />
                </Button>
              }
            />
            <DropdownMenuContent align="end" className="w-48">
              {isThreadRoot && canEdit && (
                <DropdownMenuItem
                  onClick={() =>
                    setResolved({ commentId: comment.id, resolved: !comment.resolved_at })
                  }
                >
                  {comment.resolved_at ? (
                    <UndoIcon className="size-4" />
                  ) : (
                    <CheckCircle2Icon className="size-4" />
                  )}
                  {comment.resolved_at ? "Unresolve thread" : "Resolve thread"}
                </DropdownMenuItem>
              )}
              {canModify && (
                <DropdownMenuItem
                  onClick={() => {
                    setDraft(comment.body);
                    setIsEditing(true);
                  }}
                >
                  <PencilIcon className="size-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {canModify && (
                <DropdownMenuItem className="text-destructive" onClick={() => setIsDeleting(true)}>
                  <Trash2Icon className="size-4" />
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </header>
      {isEditing ? (
        <div className="flex flex-col gap-2">
          <MarkdownEditor
            label="Edit comment"
            value={draft}
            onChange={setDraft}
            onSubmit={() => void save().catch(() => undefined)}
            users={users}
            onUploadImage={uploadImage}
            autoFocus
          />
          <div className="flex items-center justify-end gap-2">
            <Button variant="outline" size="sm" onClick={() => setIsEditing(false)}>
              Cancel
            </Button>
            <Button
              size="sm"
              isPending={isSaving}
              disabled={!draft.trim()}
              onClick={() => void save().catch(() => undefined)}
            >
              Save
            </Button>
          </div>
        </div>
      ) : (
        <MarkdownView markdown={comment.body} />
      )}
      <ReactionBar
        reactions={comment.reactions}
        disabled={!canEdit}
        onToggle={(emoji, reacted) => toggleReaction({ commentId: comment.id, emoji, reacted })}
      />
      <ConfirmDialog
        open={isDeleting}
        onOpenChange={setIsDeleting}
        title="Delete comment"
        description={
          "replies" in comment && comment.replies.length > 0
            ? "The comment and its replies are deleted for everyone."
            : "The comment is deleted for everyone."
        }
        confirmLabel="Delete comment"
        onConfirm={() => deleteComment({ commentId: comment.id })}
      />
    </article>
  );
}
