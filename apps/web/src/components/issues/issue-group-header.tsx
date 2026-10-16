import { ChevronRightIcon, PlusIcon } from "lucide-react";

import { GroupIcon } from "@/components/issues/group-icon";
import type { TIssueGroup } from "@/components/issues/grouping";
import { Button } from "@/components/ui/button";

type TProps = {
  group: TIssueGroup;
  isCollapsed: boolean;
  onToggle: () => void;
  onCreate?: () => void;
};

export function IssueGroupHeader({ group, isCollapsed, onToggle, onCreate }: TProps) {
  return (
    <div
      role="listitem"
      aria-label={`${group.label}, ${group.issues.length} issues`}
      className="bg-muted flex h-9 w-full items-center gap-2 border-b px-1.5 text-sm"
    >
      <Button
        variant="ghost"
        size="sm"
        forceMinSize={false}
        aria-expanded={!isCollapsed}
        onClick={onToggle}
        className="min-w-0 justify-start"
      >
        <ChevronRightIcon
          data-open={!isCollapsed || undefined}
          className="text-muted-foreground size-3.5 transition-transform data-open:rotate-90"
        />
        <GroupIcon meta={group.meta} />
        <span className="truncate">{group.label}</span>
        <span className="text-muted-foreground text-xs font-normal">{group.issues.length}</span>
      </Button>
      <span className="flex-1" />
      {onCreate && (
        <Button
          variant="ghost"
          size="icon"
          aria-label={`New issue in ${group.label}`}
          onClick={onCreate}
        >
          <PlusIcon className="size-4" />
        </Button>
      )}
    </div>
  );
}
