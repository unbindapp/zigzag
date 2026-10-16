import {
  ArchiveIcon,
  ArchiveRestoreIcon,
  EllipsisIcon,
  PencilIcon,
  PlusIcon,
  Trash2Icon,
} from "lucide-react";
import { useState } from "react";

import { LabelDialog } from "@/components/labels/label-dialog";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "@/components/ui/toast";
import {
  type TLabel,
  type TLabelTree,
  useArchiveLabelMutation,
  useDeleteLabelMutation,
  useUnarchiveLabelMutation,
} from "@/lib/queries/labels";

type TOpenDialog = "edit" | "add" | "delete" | null;

export function LabelRowMenu({ label, groups }: { label: TLabel; groups: TLabelTree[] }) {
  const [dialog, setDialog] = useState<TOpenDialog>(null);
  const { mutate: archive } = useArchiveLabelMutation();
  const { mutate: unarchive } = useUnarchiveLabelMutation();
  const { mutateAsync: deleteLabel } = useDeleteLabelMutation();
  const noun = label.is_group ? "group" : "label";
  const close = (open: boolean) => {
    if (!open) setDialog(null);
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon" aria-label={`Actions for ${label.name}`}>
              <EllipsisIcon className="size-4" />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="w-48">
          <DropdownMenuItem onClick={() => setDialog("edit")}>
            <PencilIcon className="size-4" />
            Edit {noun}
          </DropdownMenuItem>
          {label.is_group && !label.archived_at && (
            <DropdownMenuItem onClick={() => setDialog("add")}>
              <PlusIcon className="size-4" />
              Add label to group
            </DropdownMenuItem>
          )}
          {label.archived_at ? (
            <DropdownMenuItem
              onClick={() =>
                unarchive(
                  { labelId: label.id },
                  {
                    onSuccess: () =>
                      toast.add({ type: "success", title: `Restored ${label.name}` }),
                  },
                )
              }
            >
              <ArchiveRestoreIcon className="size-4" />
              Unarchive
            </DropdownMenuItem>
          ) : (
            <DropdownMenuItem
              onClick={() =>
                archive(
                  { labelId: label.id },
                  {
                    onSuccess: () =>
                      toast.add({ type: "success", title: `Archived ${label.name}` }),
                  },
                )
              }
            >
              <ArchiveIcon className="size-4" />
              Archive
            </DropdownMenuItem>
          )}
          <DropdownMenuSeparator />
          <DropdownMenuItem className="text-destructive" onClick={() => setDialog("delete")}>
            <Trash2Icon className="size-4" />
            Delete {noun}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <LabelDialog
        mode="edit"
        label={label}
        groups={groups}
        open={dialog === "edit"}
        onOpenChange={close}
      />
      {label.is_group && (
        <LabelDialog
          mode="create"
          workspaceId={label.workspace_id}
          teamId={label.team_id ?? undefined}
          isGroup={false}
          parentId={label.id}
          groups={groups}
          open={dialog === "add"}
          onOpenChange={close}
        />
      )}
      <ConfirmDialog
        open={dialog === "delete"}
        onOpenChange={close}
        title={`Delete ${label.name}`}
        description={
          label.is_group
            ? "Deletes the group and every label in it, and removes them from all issues."
            : "Deletes the label and removes it from all issues. Archive it instead to keep it on existing issues."
        }
        confirmLabel={`Delete ${noun}`}
        onConfirm={() => deleteLabel({ labelId: label.id })}
      />
    </>
  );
}
