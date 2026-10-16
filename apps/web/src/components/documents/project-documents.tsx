import { PlusIcon } from "lucide-react";

import { DocumentList } from "@/components/documents/document-list";
import { useCreateDocument } from "@/components/documents/use-create-document";
import { SectionTitle } from "@/components/layout/section-title";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";

type TProps = {
  workspaceId: string;
  projectId?: string;
  initiativeId?: string;
};

export function ProjectDocuments({ workspaceId, projectId, initiativeId }: TProps) {
  const { canEdit } = useWorkspace();
  const { create, isPending } = useCreateDocument();
  return (
    <section aria-label="Documents" className="flex w-full flex-col gap-4">
      <SectionTitle
        title="Documents"
        actions={
          canEdit && (
            <Button
              size="sm"
              variant="ghost"
              isPending={isPending}
              onClick={() => create({ projectId, initiativeId })}
            >
              <PlusIcon className="size-4" />
              New document
            </Button>
          )
        }
      />
      <DocumentList
        workspaceId={workspaceId}
        projectId={projectId}
        initiativeId={initiativeId}
        skeletonRows={2}
      />
    </section>
  );
}
