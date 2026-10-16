import { PlusIcon } from "lucide-react";

import { useIsPeekable, useIssuePeek } from "@/components/issues/issue-peek";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { AssigneePicker, StatePicker } from "@/components/issues/pickers";
import type { TIssueDetail } from "@/lib/queries/issues";
import { useIssuePatch, useOpenIssuePage } from "@/components/issues/use-issue-actions";
import { List } from "@/components/layout/list";
import { SectionTitle } from "@/components/layout/section-title";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export function SubIssues({ issue }: { issue: TIssueDetail }) {
  const { canEdit } = useWorkspace();
  const openNewIssue = useNewIssueStore((state) => state.open);
  const patch = useIssuePatch();
  const peek = useIssuePeek();
  const isPeekable = useIsPeekable();
  const openPage = useOpenIssuePage();
  const children = [...(issue.children ?? [])].sort(
    (a, b) => a.sub_issue_sort_order - b.sub_issue_sort_order,
  );
  const completed = children.filter(
    (child) => child.state.type === "completed" || child.state.type === "canceled",
  ).length;

  const addSubIssue = () =>
    openNewIssue({
      teamId: issue.team_id,
      parentId: issue.id,
      draft: {
        parent: { id: issue.id, identifier: issue.identifier, title: issue.title },
        project: issue.project ?? null,
        cycle: issue.cycle ?? null,
      },
    });

  const open = (identifier: string) => {
    if (peek && isPeekable) {
      peek.openPeek(identifier);
      return;
    }
    void openPage(identifier);
  };

  return (
    <section aria-label="Sub-issues" className="flex w-full flex-col gap-2">
      <SectionTitle
        title={children.length > 0 ? `Sub-issues ${completed}/${children.length}` : "Sub-issues"}
        actions={
          canEdit && (
            <Button variant="ghost" size="sm" onClick={addSubIssue}>
              <PlusIcon className="size-4" />
              Add sub-issue
            </Button>
          )
        }
      />
      {children.length > 0 && (
        <List>
          {children.map((child) => (
            <div
              key={child.id}
              role="listitem"
              tabIndex={0}
              onClick={() => open(child.identifier)}
              onKeyDown={(event) =>
                event.key === "Enter" &&
                event.target === event.currentTarget &&
                open(child.identifier)
              }
              className="has-hover:hover:bg-accent focus-visible:ring-primary/8-10 flex h-9 w-full cursor-default items-center gap-2 border-b px-3 text-sm outline-hidden last:border-b-0 focus-visible:ring-1 focus-visible:ring-inset"
            >
              <StatePicker
                appearance="icon"
                disabled={!canEdit}
                teamId={child.team_id}
                state={child.state}
                onChange={(state) => patch(child, { state })}
              />
              <span className="text-muted-foreground w-16 shrink-0 truncate font-mono text-xs">
                {child.identifier}
              </span>
              <span className="min-w-0 flex-1 truncate">{child.title}</span>
              <AssigneePicker
                appearance="icon"
                disabled={!canEdit}
                assignee={child.assignee}
                onChange={(assignee) => patch(child, { assignee })}
              />
            </div>
          ))}
        </List>
      )}
    </section>
  );
}
