import { useNavigate } from "@tanstack/react-router";
import { EllipsisIcon, PencilIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ViewDialog } from "@/components/views/view-dialog";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TView, useDeleteViewMutation } from "@/lib/queries/views";

export function useCanEditView(view: TView) {
  const { canEdit, isAdmin } = useWorkspace();
  return view.is_owner ? canEdit : isAdmin && view.shared;
}

export function ViewMenu({ view, leaveOnDelete }: { view: TView; leaveOnDelete?: boolean }) {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const [dialog, setDialog] = useState<"edit" | "delete" | null>(null);
  const { mutateAsync: deleteView } = useDeleteViewMutation();
  const close = (open: boolean) => {
    if (!open) setDialog(null);
  };
  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon" aria-label={`Actions for ${view.name}`}>
              <EllipsisIcon className="size-4" />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="w-40">
          <DropdownMenuItem onClick={() => setDialog("edit")}>
            <PencilIcon className="size-4" />
            Edit view
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem className="text-destructive" onClick={() => setDialog("delete")}>
            <Trash2Icon className="size-4" />
            Delete view
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <ViewDialog mode="edit" view={view} open={dialog === "edit"} onOpenChange={close} />
      <ConfirmDialog
        open={dialog === "delete"}
        onOpenChange={close}
        title={`Delete ${view.name}`}
        description={
          view.shared
            ? "The view is deleted for everyone in the workspace. Issues are not touched."
            : "The view is deleted. Issues are not touched."
        }
        confirmLabel="Delete view"
        onConfirm={async () => {
          await deleteView({ viewId: view.id });
          if (leaveOnDelete) {
            void navigate({ to: "/$workspace/views", params: { workspace: workspace.slug } });
          }
        }}
      />
    </>
  );
}
