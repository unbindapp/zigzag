import { createFileRoute } from "@tanstack/react-router";
import { BoxIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/projects")({
  component: ProjectsPage,
});

function ProjectsPage() {
  return (
    <PageContainer>
      <PageHeader title={"Projects"} icon={<BoxIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
