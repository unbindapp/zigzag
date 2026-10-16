import { ReactNode, useId } from "react";

import { Checkbox } from "@/components/ui/checkbox";

export type TCheckboxListItem = {
  value: string;
  label: string;
  icon?: ReactNode;
  description?: string;
};

type TProps = {
  items: TCheckboxListItem[];
  value: string[];
  onChange: (value: string[]) => void;
  label: string;
  emptyText?: string;
};

export function CheckboxList({ items, value, onChange, label, emptyText }: TProps) {
  const groupId = useId();
  if (items.length === 0) {
    return <p className="text-muted-foreground text-sm">{emptyText ?? "Nothing to pick."}</p>;
  }
  return (
    <div
      role="group"
      aria-label={label}
      className="flex max-h-56 flex-col overflow-y-auto rounded-md border"
    >
      {items.map((item) => {
        const id = `${groupId}-${item.value}`;
        const checked = value.includes(item.value);
        return (
          <label
            key={item.value}
            htmlFor={id}
            className="has-hover:hover:bg-accent flex h-9 shrink-0 cursor-pointer items-center gap-2 border-b px-3 text-sm last:border-b-0"
          >
            <Checkbox
              id={id}
              checked={checked}
              onCheckedChange={(next) =>
                onChange(next ? [...value, item.value] : value.filter((v) => v !== item.value))
              }
            />
            {item.icon}
            <span className="min-w-0 flex-1 truncate">{item.label}</span>
            {item.description && (
              <span className="text-muted-foreground truncate text-xs">{item.description}</span>
            )}
          </label>
        );
      })}
    </div>
  );
}
