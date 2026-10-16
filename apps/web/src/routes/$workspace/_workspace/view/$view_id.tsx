import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LayersIcon, PlusIcon } from "lucide-react";
import { useState } from "react";

import { FavoriteButton } from "@/components/favorites/favorite-button";
import { compactFilters, type ListIssuesFilters, mergeFilters } from "@/components/issues/filters";
import { IssueBoard } from "@/components/issues/issue-board";
import { IssueFilters } from "@/components/issues/issue-filters";
import { IssueList } from "@/components/issues/issue-list";
import { IssuePeekHost } from "@/components/issues/issue-peek";
import { useNewIssueContext, useNewIssueStore } from "@/components/issues/new-issue-store";
import { useFilterDefaults } from "@/components/issues/use-filter-defaults";
import type { TBreadcrumbItem } from "@/components/layout/breadcrumb";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { PageToolbar } from "@/components/layout/page-toolbar";
import { EntityIcon } from "@/components/shared/entity-icon";
import { Button } from "@/components/ui/button";
import { displaySettingsOf, filtersOf, viewFiltersFrom } from "@/components/views/view-display";
import { useCanEditView, ViewMenu } from "@/components/views/view-menu";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";
import { type TView, useUpdateViewMutation, viewQuery } from "@/lib/queries/views";
import { EntityNotFound } from "@/components/navigation/not-found-template";

export const Route = createFileRoute("/$workspace/_workspace/view/$view_id")({
  component: ViewPage,
});

function ViewPage() {
  const { workspace, view_id } = Route.useParams();
  const { data, isPending, error, refetch } = useQuery(viewQuery(view_id));
  const breadcrumb: TBreadcrumbItem[] = [
    { label: "Views", link: { to: "/$workspace/views", params: { workspace } } },
  ];

  if (isPending || error) {
    return (
      <PageContainer>
        <PageHeader title="View" icon={<LayersIcon />} breadcrumb={breadcrumb} />
        {isPending && <ListSkeleton />}
        {error && isApiError(error, 404) && (
          <EntityNotFound
            Icon={LayersIcon}
            description="View not found. It was deleted, or it is private to someone else."
          />
        )}
        {error && !isApiError(error, 404) && (
          <ErrorCard message={error.message} onRetry={() => void refetch()} />
        )}
      </PageContainer>
    );
  }
  return <SavedView key={data.view.id} view={data.view} breadcrumb={breadcrumb} />;
}

function SavedView({ view, breadcrumb }: { view: TView; breadcrumb: TBreadcrumbItem[] }) {
  const { workspace, canEdit: canCreateIssues } = useWorkspace();
  const canEdit = useCanEditView(view);
  const saved = filtersOf(view);
  const [filters, setFilters] = useState<ListIssuesFilters>(saved);
  const { mutate: updateView, isPending: isSaving } = useUpdateViewMutation();
  const isDirty = JSON.stringify(compactFilters(filters)) !== JSON.stringify(saved);
  const display = displaySettingsOf(view);
  const fixedFilters = view.team_id ? { team_ids: [view.team_id] } : undefined;
  const listProps = { workspaceId: workspace.id, filters, display, fixedFilters };
  const defaults = useFilterDefaults(
    mergeFilters(filters, fixedFilters),
    view.team_id ?? undefined,
  );
  const openNewIssue = useNewIssueStore((state) => state.open);
  useNewIssueContext(defaults);

  return (
    <IssuePeekHost>
      <PageContainer>
        <PageHeader
          title={view.name}
          icon={<EntityIcon icon={view.icon} color={view.color} fallback={LayersIcon} size="lg" />}
          description={view.description || undefined}
          breadcrumb={[...breadcrumb, { label: view.name }]}
          actions={
            <>
              <FavoriteButton kind="view" entityId={view.id} />
              {canEdit && <ViewMenu view={view} leaveOnDelete />}
              {canCreateIssues && (
                <Button onClick={() => openNewIssue()}>
                  <PlusIcon className="size-4" />
                  New issue
                </Button>
              )}
            </>
          }
          toolbar={
            <PageToolbar
              actions={
                isDirty && (
                  <>
                    <Button variant="ghost" size="sm" onClick={() => setFilters(saved)}>
                      Reset
                    </Button>
                    {canEdit && (
                      <Button
                        size="sm"
                        isPending={isSaving}
                        onClick={() =>
                          updateView({ view_id: view.id, filters: viewFiltersFrom(filters) })
                        }
                      >
                        Save filters
                      </Button>
                    )}
                  </>
                )
              }
            >
              <IssueFilters
                filters={filters}
                onChange={setFilters}
                teamId={view.team_id ?? undefined}
              />
            </PageToolbar>
          }
        />
        {display.layout === "board" ? <IssueBoard {...listProps} /> : <IssueList {...listProps} />}
      </PageContainer>
    </IssuePeekHost>
  );
}
