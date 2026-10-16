import { createFileRoute } from "@tanstack/react-router";
import { ListTodoIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/issues")({
  component: IssuesPage,
});

function IssuesPage() {
  return (
    <PageContainer>
      <PageHeader title={"All issues"} icon={<ListTodoIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
