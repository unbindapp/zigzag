import { createFileRoute } from "@tanstack/react-router";
import { LayersIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/views")({
  component: ViewsPage,
});

function ViewsPage() {
  return (
    <PageContainer>
      <PageHeader title={"Views"} icon={<LayersIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
