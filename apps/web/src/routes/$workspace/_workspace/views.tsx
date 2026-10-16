import { useNavigate } from "@tanstack/react-router";
import { createFileRoute } from "@tanstack/react-router";
import { LayersIcon, PlusIcon } from "lucide-react";

import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import { ViewDialog } from "@/components/views/view-dialog";
import { defaultViewDisplay } from "@/components/views/view-display";
import { ViewList } from "@/components/views/view-list";
import { useWorkspace } from "@/components/workspace/workspace-provider";

export const Route = createFileRoute("/$workspace/_workspace/views")({
  component: ViewsPage,
});

function ViewsPage() {
  const { workspace, canEdit } = useWorkspace();
  const navigate = useNavigate();
  return (
    <PageContainer>
      <PageHeader
        title="Views"
        icon={<LayersIcon />}
        description="Saved issue filters. Private views are yours; shared views are for everyone."
        actions={
          canEdit && (
            <ViewDialog
              mode="create"
              draft={{ workspaceId: workspace.id, filters: {}, display: defaultViewDisplay }}
              onSaved={(view) =>
                void navigate({
                  to: "/$workspace/view/$view_id",
                  params: { workspace: workspace.slug, view_id: view.id },
                })
              }
              trigger={
                <Button>
                  <PlusIcon className="size-4" />
                  New view
                </Button>
              }
            />
          )
        }
      />
      <ViewList />
    </PageContainer>
  );
}
