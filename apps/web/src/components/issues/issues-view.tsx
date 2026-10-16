import { BookmarkPlusIcon, PlusIcon } from "lucide-react";
import type { ReactNode } from "react";

import { DisplayOptions, useDisplaySettings } from "@/components/issues/display-options";
import type { DisplaySettings } from "@/components/issues/display";
import { type ListIssuesFilters, mergeFilters } from "@/components/issues/filters";
import { IssueBoard } from "@/components/issues/issue-board";
import { IssueFilters } from "@/components/issues/issue-filters";
import { IssueList } from "@/components/issues/issue-list";
import { IssuePeekHost } from "@/components/issues/issue-peek";
import { useNewIssueContext, useNewIssueStore } from "@/components/issues/new-issue-store";
import { useFilterDefaults } from "@/components/issues/use-filter-defaults";
import type { TFilterField } from "@/components/issues/filters";
import type { TBreadcrumbItem } from "@/components/layout/breadcrumb";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { PageToolbar } from "@/components/layout/page-toolbar";
import { Button } from "@/components/ui/button";
import { useSaveViewDialog } from "@/components/views/save-view-dialog";
import { useWorkspace } from "@/components/workspace/workspace-provider";

type TProps = {
  title: string;
  icon: ReactNode;
  breadcrumb?: TBreadcrumbItem[];
  storageKey: string;
  filters: ListIssuesFilters;
  onFiltersChange: (filters: ListIssuesFilters) => void;
  fixedFilters?: ListIssuesFilters;
  teamId?: string;
  defaultDisplay?: Partial<DisplaySettings>;
  hiddenFilterFields?: TFilterField[];
  tabs?: ReactNode;
  emptyState?: ReactNode;
};

export function IssuesView({
  title,
  icon,
  breadcrumb,
  storageKey,
  filters,
  onFiltersChange,
  fixedFilters,
  teamId,
  defaultDisplay,
  hiddenFilterFields,
  tabs,
  emptyState,
}: TProps) {
  const { workspace, canEdit } = useWorkspace();
  const { display, setDisplay, reset, isDefault } = useDisplaySettings(storageKey, defaultDisplay);
  const effectiveFilters = mergeFilters(filters, fixedFilters);
  const defaults = useFilterDefaults(effectiveFilters, teamId);
  const { openSaveView, saveViewDialog } = useSaveViewDialog({
    filters: effectiveFilters,
    display,
    teamId,
  });
  const openNewIssue = useNewIssueStore((state) => state.open);
  useNewIssueContext(defaults);

  const listProps = {
    workspaceId: workspace.id,
    filters,
    fixedFilters,
    display,
    emptyState,
  };

  return (
    <IssuePeekHost>
      <PageContainer>
        <PageHeader
          title={title}
          icon={icon}
          breadcrumb={breadcrumb}
          actions={
            canEdit && (
              <Button onClick={() => openNewIssue()}>
                <PlusIcon className="size-4" />
                New issue
              </Button>
            )
          }
          toolbar={
            <div className="flex w-full flex-col gap-2">
              {tabs}
              <PageToolbar
                actions={
                  <>
                    {canEdit && (
                      <Button variant="ghost" size="sm" onClick={openSaveView}>
                        <BookmarkPlusIcon className="size-4" />
                        Save view
                      </Button>
                    )}
                    <DisplayOptions
                      value={display}
                      onChange={setDisplay}
                      onReset={isDefault ? undefined : reset}
                    />
                  </>
                }
              >
                <IssueFilters
                  filters={filters}
                  onChange={onFiltersChange}
                  teamId={teamId}
                  hiddenFields={hiddenFilterFields}
                />
              </PageToolbar>
            </div>
          }
        />
        {display.layout === "board" ? <IssueBoard {...listProps} /> : <IssueList {...listProps} />}
        {saveViewDialog}
      </PageContainer>
    </IssuePeekHost>
  );
}
