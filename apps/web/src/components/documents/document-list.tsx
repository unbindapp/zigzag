import { useQuery } from "@tanstack/react-query";
import { FileTextIcon } from "lucide-react";
import type { ReactNode } from "react";

import { DocumentRow } from "@/components/documents/document-row";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { documentsListQuery } from "@/lib/queries/documents";

type TProps = {
  workspaceId: string;
  projectId?: string;
  initiativeId?: string;
  emptyDescription: string;
  emptyAction?: ReactNode;
  skeletonRows?: number;
};

export function DocumentList({
  workspaceId,
  projectId,
  initiativeId,
  emptyDescription,
  emptyAction,
  skeletonRows,
}: TProps) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(
    documentsListQuery({
      workspace_id: workspaceId,
      project_id: projectId,
      initiative_id: initiativeId,
    }),
  );
  const documents = data?.documents ?? [];
  return (
    <ListStates
      isPending={isPending}
      error={error}
      onRetry={() => void refetch()}
      isEmpty={documents.length === 0}
      skeletonRows={skeletonRows}
      empty={{
        icon: FileTextIcon,
        title: "No documents yet",
        description: emptyDescription,
        action: emptyAction,
      }}
    >
      <List>
        {documents.map((document) => (
          <DocumentRow key={document.id} workspaceSlug={workspace.slug} document={document} />
        ))}
      </List>
    </ListStates>
  );
}
