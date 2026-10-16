import { createFileRoute } from "@tanstack/react-router";
import { CircleDotIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/issue/$identifier")({
  component: IssueIdentifierPage,
});

function IssueIdentifierPage() {
  const { workspace, identifier } = Route.useParams();
  return (
    <PageContainer>
      <PageHeader
        title={identifier}
        icon={<CircleDotIcon />}
        breadcrumb={[
          { label: "All issues", link: { to: "/$workspace/issues", params: { workspace } } },
          { label: identifier },
        ]}
      />
      <NotBuiltYet />
    </PageContainer>
  );
}
