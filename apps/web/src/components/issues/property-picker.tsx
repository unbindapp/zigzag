import { CheckIcon } from "lucide-react";
import { type ReactElement, type ReactNode, useState } from "react";

import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ScrollArea } from "@/components/ui/scroll-area";

export type TPickerOption = {
  value: string;
  label: string;
  icon?: ReactNode;
  description?: string;
  keywords?: string[];
  group?: string;
  shortcut?: string;
};

type TProps = {
  label: string;
  options: TPickerOption[];
  value: string[];
  onChange: (value: string[]) => void;
  trigger: ReactElement;
  multiple?: boolean;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  isPending?: boolean;
  emptyText?: string;
  searchPlaceholder?: string;
  onSearchChange?: (query: string) => void;
  disabled?: boolean;
  align?: "start" | "center" | "end";
  children?: ReactNode;
};

export function PropertyPicker({
  label,
  options,
  value,
  onChange,
  trigger,
  multiple = false,
  open,
  onOpenChange,
  isPending,
  emptyText = "No results",
  searchPlaceholder,
  onSearchChange,
  disabled,
  align = "start",
  children,
}: TProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = open ?? internalOpen;
  const setOpen = (next: boolean) => {
    setInternalOpen(next);
    onOpenChange?.(next);
  };

  const select = (optionValue: string) => {
    if (!multiple) {
      onChange([optionValue]);
      setOpen(false);
      return;
    }
    const isSelected = value.includes(optionValue);
    onChange(isSelected ? value.filter((v) => v !== optionValue) : [...value, optionValue]);
  };

  return (
    <span className="contents" onClick={(event) => event.stopPropagation()}>
      <Popover open={isOpen} onOpenChange={setOpen}>
        <PopoverTrigger disabled={disabled} render={trigger} />
        <PopoverContent align={align} className="w-64 p-0">
          {isOpen && (
            <Command wrapper="none" loop shouldFilter={!onSearchChange} label={label}>
              <CommandInput
                autoFocus
                showSpinner={isPending}
                placeholder={searchPlaceholder ?? `${label}…`}
                onValueChange={onSearchChange}
              />
              <ScrollArea className="max-h-72 min-h-0">
                <CommandList>
                  <CommandEmpty>{isPending ? "Loading…" : emptyText}</CommandEmpty>
                  <PickerGroups
                    options={options}
                    value={value}
                    multiple={multiple}
                    onSelect={select}
                  />
                </CommandList>
              </ScrollArea>
              {children}
            </Command>
          )}
        </PopoverContent>
      </Popover>
    </span>
  );
}

function PickerGroups({
  options,
  value,
  multiple,
  onSelect,
}: {
  options: TPickerOption[];
  value: string[];
  multiple: boolean;
  onSelect: (value: string) => void;
}) {
  const groups = new Map<string, TPickerOption[]>();
  for (const option of options) {
    const key = option.group ?? "";
    groups.set(key, [...(groups.get(key) ?? []), option]);
  }
  return [...groups.entries()].map(([heading, items]) => (
    <CommandGroup key={heading || "options"} heading={heading || undefined}>
      {items.map((option) => {
        const isSelected = value.includes(option.value);
        return (
          <CommandItem
            key={option.value}
            value={option.value}
            keywords={[option.label, ...(option.keywords ?? [])]}
            onSelect={() => onSelect(option.value)}
            data-checked={isSelected || undefined}
          >
            {option.icon}
            <span className="min-w-0 flex-1 truncate">{option.label}</span>
            {option.description && (
              <span className="text-muted-foreground truncate text-xs">{option.description}</span>
            )}
            {option.shortcut && !isSelected && <CommandShortcut>{option.shortcut}</CommandShortcut>}
            {isSelected && (
              <CheckIcon
                aria-label={multiple ? "Selected" : "Current"}
                className="text-foreground ml-auto size-4"
              />
            )}
          </CommandItem>
        );
      })}
    </CommandGroup>
  ));
}
