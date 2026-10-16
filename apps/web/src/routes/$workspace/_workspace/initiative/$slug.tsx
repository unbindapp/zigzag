import { createFileRoute } from "@tanstack/react-router";
import { TargetIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/initiative/$slug")({
  component: InitiativeSlugPage,
});

function InitiativeSlugPage() {
  const { workspace, slug } = Route.useParams();
  return (
    <PageContainer>
      <PageHeader
        title={slug}
        icon={<TargetIcon />}
        breadcrumb={[
          { label: "Initiatives", link: { to: "/$workspace/initiatives", params: { workspace } } },
          { label: slug },
        ]}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
