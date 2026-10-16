import { createFileRoute } from "@tanstack/react-router";

import { InitiativeDetail } from "@/components/initiatives/initiative-detail";
import { PageContainer } from "@/components/layout/page-container";

export const Route = createFileRoute("/$workspace/_workspace/initiative/$slug")({
  component: InitiativeSlugPage,
});

function InitiativeSlugPage() {
  const { slug } = Route.useParams();
  return (
    <PageContainer>
      <InitiativeDetail key={slug} slug={slug} />
    </PageContainer>
  );
}
