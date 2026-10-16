import { Trash2Icon } from "lucide-react";
import { useState } from "react";

import { ProjectStatusIcon } from "@/components/projects/project-status-icon";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { type TProjectStatus, useDeleteProjectStatusMutation } from "@/lib/queries/projects";

type TProps = {
  status: TProjectStatus;
  statuses: TProjectStatus[];
};

export function DeleteProjectStatusDialog({ status, statuses }: TProps) {
  const others = statuses.filter((candidate) => candidate.id !== status.id);
  const fallback = others.find((candidate) => candidate.type === status.type) ?? others[0];
  const [targetId, setTargetId] = useState<string | null>(fallback?.id ?? null);
  const { mutateAsync: deleteStatus } = useDeleteProjectStatusMutation();
  const items = others.map((candidate) => ({ value: candidate.id, label: candidate.name }));

  return (
    <ConfirmDialog
      trigger={
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Delete ${status.name}`}
          disabled={others.length === 0}
        >
          <Trash2Icon className="size-4" />
        </Button>
      }
      title={`Delete ${status.name}`}
      description="Projects in this status move to the status you pick."
      confirmLabel="Delete status"
      confirmDisabled={!targetId}
      onConfirm={async () => {
        if (!targetId) return;
        await deleteStatus({ statusId: status.id, moveToStatusId: targetId });
      }}
    >
      <div className="flex flex-col gap-1.5">
        <p className="text-sm font-medium">Move projects to</p>
        <Select items={items} value={targetId} onValueChange={(value) => setTargetId(value)}>
          <SelectTrigger aria-label="Move projects to">
            <SelectValue placeholder="Pick a status" />
          </SelectTrigger>
          <SelectContent>
            {others.map((candidate) => (
              <SelectItem key={candidate.id} value={candidate.id}>
                <ProjectStatusIcon type={candidate.type} color={candidate.color} />
                <span className="truncate">{candidate.name}</span>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </ConfirmDialog>
  );
}
