import { TargetIcon } from "lucide-react";

import { EntityIcon } from "@/components/shared/entity-icon";

type TProps = {
  initiative: { name: string; icon: string; color: string };
};

export function InitiativeChip({ initiative }: TProps) {
  return (
    <span className="bg-background text-foreground inline-flex h-5 max-w-40 shrink-0 items-center gap-1.5 rounded-full border px-2 text-xs leading-none">
      <EntityIcon icon={initiative.icon} color={initiative.color} fallback={TargetIcon} size="sm" />
      <span className="truncate">{initiative.name}</span>
    </span>
  );
}
