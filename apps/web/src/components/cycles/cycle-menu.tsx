import { useNavigate } from "@tanstack/react-router";
import { ArchiveIcon, EllipsisIcon, PencilIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import { CycleDialog } from "@/components/cycles/cycle-dialog";
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
  cycleName,
  type TCycle,
  useArchiveCycleMutation,
  useDeleteCycleMutation,
} from "@/lib/queries/cycles";

type TProps = {
  cycle: TCycle;
  workspaceSlug: string;
  teamKey: string;
};

export function CycleMenu({ cycle, workspaceSlug, teamKey }: TProps) {
  const [dialog, setDialog] = useState<"edit" | "delete" | null>(null);
  const navigate = useNavigate();
  const { mutate: archive } = useArchiveCycleMutation();
  const { mutateAsync: deleteCycle } = useDeleteCycleMutation();
  const name = cycleName(cycle);
  const close = (open: boolean) => {
    if (!open) setDialog(null);
  };
  const backToList = () =>
    void navigate({
      to: "/$workspace/team/$key/cycles",
      params: { workspace: workspaceSlug, key: teamKey },
    });

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon" aria-label={`Actions for ${name}`}>
              <EllipsisIcon className="size-4" />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="w-44">
          <DropdownMenuItem onClick={() => setDialog("edit")}>
            <PencilIcon className="size-4" />
            Edit cycle
          </DropdownMenuItem>
          {!cycle.archived_at && (
            <DropdownMenuItem
              onClick={() =>
                archive(
                  { cycleId: cycle.id },
                  {
                    onSuccess: () => {
                      toast.add({ type: "success", title: `Archived ${name}` });
                      backToList();
                    },
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
            Delete cycle
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <CycleDialog mode="edit" cycle={cycle} open={dialog === "edit"} onOpenChange={close} />
      <ConfirmDialog
        open={dialog === "delete"}
        onOpenChange={close}
        title={`Delete ${name}`}
        description="The cycle is deleted. Its issues stay and leave the cycle."
        confirmLabel="Delete cycle"
        onConfirm={async () => {
          await deleteCycle({ cycleId: cycle.id });
          backToList();
        }}
      />
    </>
  );
}
