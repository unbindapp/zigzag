import { createFileRoute } from "@tanstack/react-router";
import { LayersIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/view/$view_id")({
  component: ViewView_idPage,
});

function ViewView_idPage() {
  const { workspace, view_id } = Route.useParams();
  return (
    <PageContainer>
      <PageHeader
        title={view_id}
        icon={<LayersIcon />}
        breadcrumb={[
          { label: "Views", link: { to: "/$workspace/views", params: { workspace } } },
          { label: view_id },
        ]}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
