import { Link } from "@tanstack/react-router";
import { FileTextIcon } from "lucide-react";

import { InitiativeChip } from "@/components/initiatives/initiative-chip";
import { userName } from "@/components/initiatives/initiative-pickers";
import { ProjectChip } from "@/components/issues/issue-chips";
import { ListRow } from "@/components/layout/list";
import { Avatar } from "@/components/shared/avatar";
import { EntityIcon } from "@/components/shared/entity-icon";
import { RelativeTime } from "@/components/shared/relative-time";
import { SimpleTooltip } from "@/components/ui/tooltip";
import type { TDocumentListItem } from "@/lib/queries/documents";

type TProps = {
  workspaceSlug: string;
  document: TDocumentListItem;
};

export function DocumentRow({ workspaceSlug, document }: TProps) {
  const editor = document.updated_by ?? document.creator;
  return (
    <ListRow muted={Boolean(document.archived_at || document.trashed_at)}>
      <EntityIcon icon={document.icon} fallback={FileTextIcon} />
      <Link
        to="/$workspace/document/$slug"
        params={{ workspace: workspaceSlug, slug: document.slug }}
        className="min-w-0 flex-1 truncate font-medium"
      >
        {document.title}
      </Link>
      <span className="hidden shrink-0 sm:flex">
        {document.project && <ProjectChip project={document.project} />}
        {!document.project && document.initiative && (
          <InitiativeChip initiative={document.initiative} />
        )}
      </span>
      <span className="flex w-6 shrink-0 justify-center">
        {editor && (
          <SimpleTooltip content={`Last edited by ${userName(editor)}`}>
            <span className="flex">
              <Avatar id={editor.id} name={userName(editor)} src={editor.avatar_url} size="sm" />
            </span>
          </SimpleTooltip>
        )}
      </span>
      <span className="text-muted-foreground w-16 shrink-0 text-right text-xs">
        <RelativeTime value={document.updated_at} />
      </span>
    </ListRow>
  );
}
