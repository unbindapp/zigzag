import { useNavigate } from "@tanstack/react-router";
import { type ReactNode, useState } from "react";

import type { DisplaySettings } from "@/components/issues/display";
import type { ListIssuesFilters } from "@/components/issues/filters";
import { ViewDialog } from "@/components/views/view-dialog";
import { viewDisplayFrom, viewFiltersFrom } from "@/components/views/view-display";
import { useWorkspace } from "@/components/workspace/workspace-provider";

type TSaveViewArgs = {
  filters: ListIssuesFilters;
  display?: Partial<DisplaySettings>;
  teamId?: string;
};

export function useSaveViewDialog({ filters, display, teamId }: TSaveViewArgs): {
  openSaveView: () => void;
  saveViewDialog: ReactNode;
} {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const saveViewDialog = (
    <ViewDialog
      mode="create"
      open={isOpen}
      onOpenChange={setIsOpen}
      draft={{
        workspaceId: workspace.id,
        teamId,
        filters: viewFiltersFrom(filters),
        display: viewDisplayFrom(display),
      }}
      onSaved={(view) =>
        void navigate({
          to: "/$workspace/view/$view_id",
          params: { workspace: workspace.slug, view_id: view.id },
        })
      }
    />
  );
  return { openSaveView: () => setIsOpen(true), saveViewDialog };
}
