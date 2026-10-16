import { type ReactNode } from "react";

import { HealthBadge } from "@/components/projects/health-badge";
import { Avatar } from "@/components/shared/avatar";
import { MarkdownView } from "@/components/shared/markdown-view";
import { RelativeTime } from "@/components/shared/relative-time";
import { type TUserSummary } from "@/components/shared/user-chip";
import { Card } from "@/components/ui/card";
import type { THealth } from "@/lib/queries/projects";

type TProps = {
  author?: TUserSummary | null;
  body: string;
  health: THealth;
  createdAt: string;
  editedAt?: string | null;
  actions?: ReactNode;
  footer?: ReactNode;
  children?: ReactNode;
};

export function StatusUpdateCard({
  author,
  body,
  health,
  createdAt,
  editedAt,
  actions,
  footer,
  children,
}: TProps) {
  const name = author?.name || author?.display_name || "Someone";
  return (
    <Card className="flex flex-col gap-3 p-4">
      <div className="flex min-w-0 items-center gap-2">
        <HealthBadge health={health} />
        <span className="inline-flex min-w-0 items-center gap-1.5 text-sm">
          {author && <Avatar id={author.id} name={name} src={author.avatar_url} size="sm" />}
          <span className="truncate font-medium">{name}</span>
        </span>
        <span className="text-muted-foreground shrink-0 text-xs">
          <RelativeTime value={createdAt} />
          {editedAt && " · edited"}
        </span>
        {actions && <div className="ml-auto flex shrink-0 items-center gap-2">{actions}</div>}
      </div>
      {children ?? <MarkdownView markdown={body} />}
      {footer}
    </Card>
  );
}
