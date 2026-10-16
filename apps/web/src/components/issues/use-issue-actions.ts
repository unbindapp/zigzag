import { useNavigate } from "@tanstack/react-router";

import { patchToOptimistic, type TIssuePatch } from "@/components/issues/issue-draft";
import {
  issuePatchToUpdateInput,
  type TIssue,
  useIssueLifecycleMutation,
  useUpdateIssueMutation,
} from "@/lib/queries/issues";
import { toast } from "@/components/ui/toast";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useCopyToClipboard } from "@/lib/hooks/use-copy";

export function useIssuePatch() {
  const { mutate } = useUpdateIssueMutation();
  return (issue: Pick<TIssue, "id">, patch: TIssuePatch) =>
    mutate({
      input: issuePatchToUpdateInput(issue.id, patch),
      optimistic: patchToOptimistic(patch),
    });
}

export function useCopyIssue() {
  const { copyToClipboard } = useCopyToClipboard();
  const copy = (text: string, title: string) => {
    copyToClipboard(text);
    toast.add({ type: "success", title, description: text });
  };
  return {
    copyLink: (issue: Pick<TIssue, "url">) => copy(issue.url, "Link copied"),
    copyBranch: (issue: Pick<TIssue, "git_branch_name">) =>
      copy(issue.git_branch_name, "Branch name copied"),
  };
}

const lifecycleToasts = {
  archive: "Issue archived",
  unarchive: "Issue unarchived",
  trash: "Issue moved to trash",
  restore: "Issue restored",
} as const;

export function useIssueLifecycle() {
  const { mutateAsync, isPending } = useIssueLifecycleMutation();
  const run = async (issueIds: string[], action: keyof typeof lifecycleToasts) => {
    await mutateAsync({ issueIds, action });
    const count = issueIds.length;
    toast.add({
      type: "success",
      title:
        count === 1
          ? lifecycleToasts[action]
          : `${count} ${lifecycleToasts[action].replace("Issue", "issues")}`,
    });
  };
  return { run, isPending };
}

export function useOpenIssuePage() {
  const navigate = useNavigate();
  const { workspace } = useWorkspace();
  return (identifier: string) =>
    navigate({
      to: "/$workspace/issue/$identifier",
      params: { workspace: workspace.slug, identifier },
    });
}
