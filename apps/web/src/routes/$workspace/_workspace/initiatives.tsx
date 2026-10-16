import { createFileRoute } from "@tanstack/react-router";
import { TargetIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/initiatives")({
  component: InitiativesPage,
});

function InitiativesPage() {
  return (
    <PageContainer>
      <PageHeader title={"Initiatives"} icon={<TargetIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
