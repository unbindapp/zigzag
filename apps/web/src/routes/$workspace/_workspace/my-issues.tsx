import { createFileRoute } from "@tanstack/react-router";
import { UserCircleIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/my-issues")({
  component: MyIssuesPage,
});

function MyIssuesPage() {
  return (
    <PageContainer>
      <PageHeader title={"My issues"} icon={<UserCircleIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
