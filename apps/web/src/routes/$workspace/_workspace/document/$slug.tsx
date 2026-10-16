import { createFileRoute } from "@tanstack/react-router";

import { DocumentDetail } from "@/components/documents/document-detail";
import { PageContainer } from "@/components/layout/page-container";

export const Route = createFileRoute("/$workspace/_workspace/document/$slug")({
  component: DocumentSlugPage,
});

function DocumentSlugPage() {
  const { slug } = Route.useParams();
  return (
    <PageContainer width="narrow">
      <DocumentDetail key={slug} slug={slug} />
    </PageContainer>
  );
}
