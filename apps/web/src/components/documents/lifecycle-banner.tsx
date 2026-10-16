import { ArchiveIcon, Trash2Icon } from "lucide-react";

import { Button } from "@/components/ui/button";

type TProps = {
  entity: string;
  state: "archived" | "trashed" | null;
  isPending?: boolean;
  onRestore?: () => void;
};

export function LifecycleBanner({ entity, state, isPending, onRestore }: TProps) {
  if (!state) return null;
  const isTrashed = state === "trashed";
  const Icon = isTrashed ? Trash2Icon : ArchiveIcon;
  return (
    <div role="status" className="bg-muted flex items-center gap-2 rounded-lg border p-4 text-sm">
      <Icon className="text-muted-foreground size-4 shrink-0" />
      <p className="min-w-0 flex-1">
        {isTrashed ? `This ${entity} is in the trash.` : `This ${entity} is archived.`}
      </p>
      {onRestore && (
        <Button variant="outline" size="sm" isPending={isPending} onClick={onRestore}>
          {isTrashed ? "Restore" : "Unarchive"}
        </Button>
      )}
    </div>
  );
}
