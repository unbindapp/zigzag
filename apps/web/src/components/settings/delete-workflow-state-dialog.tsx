import { Trash2Icon } from "lucide-react";
import { useState } from "react";

import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { StateIcon } from "@/components/shared/state-icon";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { type TWorkflowState, useDeleteWorkflowStateMutation } from "@/lib/queries/teams";

type TProps = {
  state: TWorkflowState;
  states: TWorkflowState[];
};

export function DeleteWorkflowStateDialog({ state, states }: TProps) {
  const others = states.filter((candidate) => candidate.id !== state.id);
  const fallback = others.find((candidate) => candidate.type === state.type) ?? others[0];
  const [targetId, setTargetId] = useState<string | null>(fallback?.id ?? null);
  const { mutateAsync: deleteState } = useDeleteWorkflowStateMutation();
  const items = others.map((candidate) => ({ value: candidate.id, label: candidate.name }));

  return (
    <ConfirmDialog
      trigger={
        <Button
          variant="ghost"
          size="icon"
          aria-label={`Delete ${state.name}`}
          disabled={others.length === 0}
        >
          <Trash2Icon className="size-4" />
        </Button>
      }
      title={`Delete ${state.name}`}
      description="Issues in this state move to the state you pick."
      confirmLabel="Delete state"
      confirmDisabled={!targetId}
      onConfirm={async () => {
        if (!targetId) return;
        await deleteState({ stateId: state.id, moveIssuesToStateId: targetId });
      }}
    >
      <div className="flex flex-col gap-1.5">
        <p className="text-sm font-medium">Move issues to</p>
        <Select items={items} value={targetId} onValueChange={(value) => setTargetId(value)}>
          <SelectTrigger aria-label="Move issues to">
            <SelectValue placeholder="Pick a state" />
          </SelectTrigger>
          <SelectContent>
            {others.map((candidate) => (
              <SelectItem key={candidate.id} value={candidate.id}>
                <StateIcon type={candidate.type} color={candidate.color} />
                <span className="truncate">{candidate.name}</span>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </ConfirmDialog>
  );
}
