import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { PlusIcon, TargetIcon } from "lucide-react";

import { InitiativeRow } from "@/components/initiatives/initiative-row";
import { NewInitiativeDialog } from "@/components/initiatives/new-initiative-dialog";
import { List } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { initiativesListQuery } from "@/lib/queries/initiatives";

export const Route = createFileRoute("/$workspace/_workspace/initiatives")({
  component: InitiativesPage,
});

function NewInitiativeButton({ size }: { size: "sm" | "default" }) {
  return (
    <NewInitiativeDialog
      trigger={
        <Button size={size}>
          <PlusIcon className="size-4" />
          New initiative
        </Button>
      }
    />
  );
}

function InitiativesPage() {
  const { workspace, canEdit } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(
    initiativesListQuery({ workspace_id: workspace.id }),
  );
  const initiatives = data?.initiatives ?? [];
  return (
    <PageContainer>
      <PageHeader
        title="Initiatives"
        icon={<TargetIcon />}
        actions={canEdit && <NewInitiativeButton size="default" />}
      />
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={() => void refetch()}
        isEmpty={initiatives.length === 0}
        empty={{
          icon: TargetIcon,
          title: "No initiatives yet",
          description: "Initiatives group projects toward a larger goal and track their progress.",
          action: canEdit && <NewInitiativeButton size="sm" />,
        }}
      >
        <List>
          {initiatives.map((initiative) => (
            <InitiativeRow
              key={initiative.id}
              workspaceSlug={workspace.slug}
              initiative={initiative}
            />
          ))}
        </List>
      </ListStates>
    </PageContainer>
  );
}
