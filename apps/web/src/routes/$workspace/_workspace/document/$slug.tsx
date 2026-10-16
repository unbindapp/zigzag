import { createFileRoute } from "@tanstack/react-router";
import { FileTextIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/document/$slug")({
  component: DocumentSlugPage,
});

function DocumentSlugPage() {
  const { workspace, slug } = Route.useParams();
  return (
    <PageContainer>
      <PageHeader
        title={slug}
        icon={<FileTextIcon />}
        breadcrumb={[
          { label: "Documents", link: { to: "/$workspace/documents", params: { workspace } } },
          { label: slug },
        ]}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
