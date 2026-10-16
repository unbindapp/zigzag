import { createFileRoute } from "@tanstack/react-router";
import { FileTextIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/documents")({
  component: DocumentsPage,
});

function DocumentsPage() {
  return (
    <PageContainer>
      <PageHeader title={"Documents"} icon={<FileTextIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
