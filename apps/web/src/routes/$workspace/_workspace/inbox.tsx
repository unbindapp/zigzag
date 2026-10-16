import { createFileRoute } from "@tanstack/react-router";
import { InboxIcon } from "lucide-react";

import { NotBuiltYet } from "@/components/layout/not-built-yet";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";

export const Route = createFileRoute("/$workspace/_workspace/inbox")({
  component: InboxPage,
});

function InboxPage() {
  return (
    <PageContainer>
      <PageHeader title={"Inbox"} icon={<InboxIcon />} />
      <NotBuiltYet />
    </PageContainer>
  );
}
