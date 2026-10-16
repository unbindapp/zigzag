import { createFileRoute } from "@tanstack/react-router";
import { BoxIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/project/$slug/")({
  component: ProjectSlugPage,
});

function ProjectSlugPage() {
  const { workspace, slug } = Route.useParams();
  return (
    <PageContainer>
      <PageHeader
        title={slug}
        icon={<BoxIcon />}
        breadcrumb={[
          { label: "Projects", link: { to: "/$workspace/projects", params: { workspace } } },
          { label: slug },
        ]}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
