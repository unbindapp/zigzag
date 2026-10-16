import { useId } from "react";

import { Checkbox } from "@/components/ui/checkbox";

export type TImportOptions = {
  restoreArchived: boolean;
  restoreTrashed: boolean;
  createMissingUsers: boolean;
};

export const DEFAULT_IMPORT_OPTIONS: TImportOptions = {
  restoreArchived: false,
  restoreTrashed: false,
  createMissingUsers: true,
};

const OPTIONS: { key: keyof TImportOptions; label: string; description: string }[] = [
  {
    key: "restoreArchived",
    label: "Restore archived",
    description: "Import archived issues and projects as active.",
  },
  {
    key: "restoreTrashed",
    label: "Restore trashed",
    description: "Import trashed issues and projects as active.",
  },
  {
    key: "createMissingUsers",
    label: "Create missing users",
    description:
      "People without an account get an inactive placeholder they can claim through an invitation.",
  },
];

type TProps = {
  value: TImportOptions;
  disabled?: boolean;
  onChange: (value: TImportOptions) => void;
};

export function ImportOptions({ value, disabled, onChange }: TProps) {
  const groupId = useId();
  return (
    <div role="group" aria-label="Import options" className="flex w-full flex-col gap-4">
      {OPTIONS.map((option) => {
        const id = `${groupId}-${option.key}`;
        return (
          <div key={option.key} className="flex w-full items-start gap-2">
            <Checkbox
              id={id}
              className="mt-0.5"
              checked={value[option.key]}
              disabled={disabled}
              aria-describedby={`${id}-description`}
              onCheckedChange={(checked) => onChange({ ...value, [option.key]: checked })}
            />
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              <label htmlFor={id} className="text-sm font-medium">
                {option.label}
              </label>
              <p id={`${id}-description`} className="text-muted-foreground text-sm">
                {option.description}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}
