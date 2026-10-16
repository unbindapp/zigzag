import type { TIssueDetail, TRelatedIssue } from "@/lib/queries/issues";
import { useIsPeekable, useIssuePeek } from "@/components/issues/issue-peek";
import { useOpenIssuePage } from "@/components/issues/use-issue-actions";
import { List } from "@/components/layout/list";
import { SectionTitle } from "@/components/layout/section-title";
import { StateIcon } from "@/components/shared/state-icon";

const RELATION_LABELS: { key: keyof NonNullable<TIssueDetail["relations"]>; label: string }[] = [
  { key: "blocked_by", label: "Blocked by" },
  { key: "blocks", label: "Blocks" },
  { key: "related", label: "Related" },
  { key: "duplicate_of", label: "Duplicate of" },
  { key: "duplicates", label: "Duplicates" },
];

export function IssueRelations({ issue }: { issue: TIssueDetail }) {
  const peek = useIssuePeek();
  const isPeekable = useIsPeekable();
  const openPage = useOpenIssuePage();
  const relations = issue.relations;
  if (!relations) return null;
  const rows = RELATION_LABELS.flatMap(({ key, label }) =>
    relations[key].map((relation: TRelatedIssue) => ({ label, relation })),
  );
  if (rows.length === 0) return null;
  const open = (identifier: string) => {
    if (peek && isPeekable) {
      peek.openPeek(identifier);
      return;
    }
    void openPage(identifier);
  };
  return (
    <section aria-label="Relations" className="flex w-full flex-col gap-2">
      <SectionTitle title="Relations" />
      <List>
        {rows.map(({ label, relation }) => (
          <div
            key={relation.relation_id}
            role="listitem"
            tabIndex={0}
            onClick={() => open(relation.issue.identifier)}
            onKeyDown={(event) =>
              event.key === "Enter" &&
              event.target === event.currentTarget &&
              open(relation.issue.identifier)
            }
            className="has-hover:hover:bg-accent focus-visible:ring-primary/8-10 flex h-9 w-full cursor-default items-center gap-2 border-b px-3 text-sm outline-hidden last:border-b-0 focus-visible:ring-1 focus-visible:ring-inset"
          >
            <span className="text-muted-foreground w-24 shrink-0 truncate text-xs">{label}</span>
            {relation.issue.state && (
              <StateIcon
                type={relation.issue.state.type}
                color={relation.issue.state.color}
                name={relation.issue.state.name}
              />
            )}
            <span className="text-muted-foreground w-16 shrink-0 truncate font-mono text-xs">
              {relation.issue.identifier}
            </span>
            <span className="min-w-0 flex-1 truncate">{relation.issue.title}</span>
          </div>
        ))}
      </List>
    </section>
  );
}
