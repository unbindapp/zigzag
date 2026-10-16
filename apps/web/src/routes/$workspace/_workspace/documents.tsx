import { createFileRoute } from "@tanstack/react-router";
import { FileTextIcon, PlusIcon } from "lucide-react";

import { DocumentList } from "@/components/documents/document-list";
import { useCreateDocument } from "@/components/documents/use-create-document";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/documents")({
  component: DocumentsPage,
});

function DocumentsPage() {
  const { workspace, canEdit } = useWorkspace();
  const { create, isPending } = useCreateDocument();
  const newDocumentButton = (placement: "header" | "empty") => (
    <Button
      size={placement === "header" ? "default" : "sm"}
      variant={placement === "header" ? "default" : "outline"}
      isPending={isPending}
      onClick={() => create()}
    >
      <PlusIcon className="size-4" />
      New document
    </Button>
  );
  return (
    <PageContainer>
      <PageHeader
        title="Documents"
        icon={<FileTextIcon />}
        actions={canEdit && newDocumentButton("header")}
      />
      <DocumentList
        workspaceId={workspace.id}
        emptyDescription="Write specs, notes and decisions next to the work they describe."
        emptyAction={canEdit && newDocumentButton("empty")}
      />
    </PageContainer>
  );
}
