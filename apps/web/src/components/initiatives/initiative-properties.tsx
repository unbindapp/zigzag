import {
  InitiativeStatusPicker,
  OwnerPicker,
  TargetDatePicker,
} from "@/components/initiatives/initiative-pickers";
import { PropertyRow } from "@/components/layout/property-row";
import { HealthBadge } from "@/components/projects/health-badge";
import { ProgressBar } from "@/components/projects/progress-bar";
import { ColorPicker } from "@/components/shared/color-picker";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  NO_OWNER_ID,
  type TInitiative,
  useUpdateInitiativeMutation,
} from "@/lib/queries/initiatives";

export function InitiativeProperties({ initiative }: { initiative: TInitiative }) {
  const { canEdit } = useWorkspace();
  const { mutate: update } = useUpdateInitiativeMutation();
  const id = initiative.id;
  return (
    <div className="flex w-full flex-col">
      <PropertyRow label="Status">
        <InitiativeStatusPicker
          status={initiative.status}
          disabled={!canEdit}
          onChange={(status) => update({ initiative_id: id, status })}
        />
      </PropertyRow>
      <PropertyRow label="Owner">
        <OwnerPicker
          owner={initiative.owner}
          disabled={!canEdit}
          onChange={(owner) => update({ initiative_id: id, owner_id: owner?.id ?? NO_OWNER_ID })}
        />
      </PropertyRow>
      <PropertyRow label="Target date">
        <TargetDatePicker
          date={initiative.target_date}
          disabled={!canEdit}
          onChange={(date) => update({ initiative_id: id, target_date: date ?? "" })}
        />
      </PropertyRow>
      <PropertyRow label="Health">
        <HealthBadge health={initiative.health} />
      </PropertyRow>
      <PropertyRow label="Progress">
        <span className="min-w-0 flex-1">
          <ProgressBar percent={initiative.progress.percent} label="Initiative progress" />
        </span>
      </PropertyRow>
      <PropertyRow label="Color">
        <ColorPicker
          value={initiative.color || null}
          disabled={!canEdit}
          label="Initiative color"
          onChange={(color) => update({ initiative_id: id, color })}
        />
      </PropertyRow>
    </div>
  );
}
