import { createFileRoute } from "@tanstack/react-router";

import { IssueDetail } from "@/components/issues/issue-detail";
import { PageContainer } from "@/components/layout/page-container";

export const Route = createFileRoute("/$workspace/_workspace/issue/$identifier")({
  component: IssueIdentifierPage,
});

function IssueIdentifierPage() {
  const { identifier } = Route.useParams();
  return (
    <PageContainer>
      <IssueDetail key={identifier} identifier={identifier} />
    </PageContainer>
  );
}
