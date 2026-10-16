import {
  ArrowDownNarrowWideIcon,
  ArrowUpNarrowWideIcon,
  LayoutGridIcon,
  ListIcon,
  SlidersHorizontalIcon,
} from "lucide-react";
import { type ReactNode, useCallback, useId } from "react";
import { useLocalStorage } from "usehooks-ts";

import {
  defaultOrderFor,
  GROUP_BY_OPTIONS,
  ISSUE_PROPERTIES,
  ORDER_BY_OPTIONS,
  resolveDisplay,
  SHOW_COMPLETED_OPTIONS,
  type TDisplaySettings,
} from "@/components/issues/display";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

const STORAGE_PREFIX = "issue-display:";

export function useDisplaySettings(storageKey: string, base?: Partial<TDisplaySettings>) {
  const [stored, setStored, removeStored] = useLocalStorage<Partial<TDisplaySettings>>(
    `${STORAGE_PREFIX}${storageKey}`,
    {},
  );
  const display = resolveDisplay(base, stored);
  const setDisplay = useCallback(
    (patch: Partial<TDisplaySettings>) => setStored((current) => ({ ...current, ...patch })),
    [setStored],
  );
  return { display, setDisplay, reset: removeStored, isDefault: Object.keys(stored).length === 0 };
}

type TProps = {
  value: TDisplaySettings;
  onChange: (patch: Partial<TDisplaySettings>) => void;
  onReset?: () => void;
};

function OptionRow({
  label,
  htmlFor,
  children,
}: {
  label: string;
  htmlFor?: string;
  children: ReactNode;
}) {
  return (
    <div className="flex items-center gap-2">
      <label htmlFor={htmlFor} className="text-muted-foreground min-w-0 flex-1 text-sm">
        {label}
      </label>
      <div className="flex shrink-0 items-center gap-2">{children}</div>
    </div>
  );
}

function OptionSelect<T extends string>({
  label,
  value,
  items,
  onChange,
}: {
  label: string;
  value: T;
  items: { value: T; label: string }[];
  onChange: (value: T) => void;
}) {
  return (
    <Select items={items} value={value} onValueChange={(next) => next && onChange(next as T)}>
      <SelectTrigger size="sm" aria-label={label} className="w-36">
        <SelectValue />
      </SelectTrigger>
      <SelectContent align="end">
        {items.map((item) => (
          <SelectItem key={item.value} value={item.value}>
            {item.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

export function DisplayOptions({ value, onChange, onReset }: TProps) {
  const subIssuesId = useId();
  const layouts = [
    { value: "list" as const, label: "List", icon: ListIcon },
    { value: "board" as const, label: "Board", icon: LayoutGridIcon },
  ];
  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button variant="outline" size="sm">
            <SlidersHorizontalIcon className="size-4" />
            Display
          </Button>
        }
      />
      <PopoverContent align="end" className="flex w-80 flex-col gap-4">
        <div role="radiogroup" aria-label="Layout" className="grid grid-cols-2 gap-2">
          {layouts.map((layout) => (
            <Button
              key={layout.value}
              role="radio"
              aria-checked={value.layout === layout.value}
              variant="outline"
              size="sm"
              data-active={value.layout === layout.value || undefined}
              className="data-active:bg-accent data-active:text-foreground"
              onClick={() => onChange({ layout: layout.value })}
            >
              <layout.icon className="size-4" />
              {layout.label}
            </Button>
          ))}
        </div>
        <div className="flex flex-col gap-2">
          <OptionRow label="Grouping">
            <OptionSelect
              label="Grouping"
              value={value.group_by}
              items={GROUP_BY_OPTIONS}
              onChange={(group_by) => onChange({ group_by })}
            />
          </OptionRow>
          <OptionRow label="Ordering">
            <OptionSelect
              label="Ordering"
              value={value.order_by}
              items={ORDER_BY_OPTIONS}
              onChange={(order_by) => onChange({ order_by, order: defaultOrderFor(order_by) })}
            />
            <Button
              variant="ghost"
              size="icon"
              aria-label={
                value.order === "asc"
                  ? "Ascending, switch to descending"
                  : "Descending, switch to ascending"
              }
              onClick={() => onChange({ order: value.order === "asc" ? "desc" : "asc" })}
            >
              {value.order === "asc" ? (
                <ArrowUpNarrowWideIcon className="size-4" />
              ) : (
                <ArrowDownNarrowWideIcon className="size-4" />
              )}
            </Button>
          </OptionRow>
          <OptionRow label="Completed issues">
            <OptionSelect
              label="Completed issues"
              value={value.show_completed}
              items={SHOW_COMPLETED_OPTIONS}
              onChange={(show_completed) => onChange({ show_completed })}
            />
          </OptionRow>
          <OptionRow label="Nest sub-issues" htmlFor={subIssuesId}>
            <Switch
              id={subIssuesId}
              checked={value.nest_sub_issues}
              onCheckedChange={(nest_sub_issues) => onChange({ nest_sub_issues })}
            />
          </OptionRow>
        </div>
        <div className="flex flex-col gap-2">
          <p className="text-muted-foreground text-xs font-medium">Display properties</p>
          <div className="flex flex-wrap gap-1.5">
            {ISSUE_PROPERTIES.map((property) => {
              const isOn = value.properties.includes(property.value);
              return (
                <Button
                  key={property.value}
                  variant="outline"
                  size="sm"
                  aria-pressed={isOn}
                  data-active={isOn || undefined}
                  className="data-active:bg-accent data-active:text-foreground text-muted-foreground"
                  onClick={() =>
                    onChange({
                      properties: isOn
                        ? value.properties.filter((item) => item !== property.value)
                        : [...value.properties, property.value],
                    })
                  }
                >
                  {property.label}
                </Button>
              );
            })}
          </div>
        </div>
        {onReset && (
          <Button variant="ghost" size="sm" className="self-start" onClick={onReset}>
            Reset to default
          </Button>
        )}
      </PopoverContent>
    </Popover>
  );
}
