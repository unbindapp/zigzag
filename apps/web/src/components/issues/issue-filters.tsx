import {
  CalendarIcon,
  CircleDashedIcon,
  CircleUserIcon,
  type LucideIcon,
  BellIcon,
  BoxIcon,
  ListFilterIcon,
  RefreshCwIcon,
  SignalHighIcon,
  TagIcon,
  UserPenIcon,
  XIcon,
} from "lucide-react";
import { useState } from "react";

import {
  activeFilterFields,
  clearFilterField,
  compactFilters,
  FILTER_FIELDS,
  listParamOfField,
  type ListIssuesFilters,
  type TFilterField,
} from "@/components/issues/filters";
import { PropertyPicker, type TPickerOption } from "@/components/issues/property-picker";
import { formatDueDate } from "@/components/issues/pickers";
import { useFilterOptions } from "@/components/issues/use-filter-options";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

const fieldIcons: Record<TFilterField, LucideIcon> = {
  status: CircleDashedIcon,
  assignee: CircleUserIcon,
  creator: UserPenIcon,
  priority: SignalHighIcon,
  labels: TagIcon,
  project: BoxIcon,
  cycle: RefreshCwIcon,
  due_date: CalendarIcon,
  subscriber: BellIcon,
};

const pluralNouns: Record<TFilterField, string> = {
  status: "statuses",
  assignee: "assignees",
  creator: "creators",
  priority: "priorities",
  labels: "labels",
  project: "projects",
  cycle: "cycles",
  due_date: "dates",
  subscriber: "subscribers",
};

type TProps = {
  filters: ListIssuesFilters;
  onChange: (filters: ListIssuesFilters) => void;
  teamId?: string;
  hiddenFields?: TFilterField[];
};

function fieldValues(filters: ListIssuesFilters, field: TFilterField): string[] {
  switch (field) {
    case "priority":
      return (filters.priorities ?? []).map(String);
    case "subscriber":
      return filters.subscriber_id ? [filters.subscriber_id] : [];
    case "due_date":
      return [];
    default: {
      const param = listParamOfField[field];
      return param ? (filters[param] ?? []) : [];
    }
  }
}

function withFieldValues(
  filters: ListIssuesFilters,
  field: TFilterField,
  values: string[],
): ListIssuesFilters {
  switch (field) {
    case "priority":
      return compactFilters({ ...filters, priorities: values.map(Number) });
    case "subscriber":
      return compactFilters({ ...filters, subscriber_id: values[values.length - 1] });
    case "due_date":
      return filters;
    default: {
      const param = listParamOfField[field];
      return param ? compactFilters({ ...filters, [param]: values }) : filters;
    }
  }
}

export function IssueFilters({ filters, onChange, teamId, hiddenFields = [] }: TProps) {
  const { options, cyclesEnabled } = useFilterOptions(teamId);
  const [pendingField, setPendingField] = useState<TFilterField | null>(null);
  const active = activeFilterFields(filters);
  const unavailable = new Set<TFilterField>([
    ...hiddenFields,
    ...(cyclesEnabled ? [] : (["cycle"] as TFilterField[])),
  ]);
  const shown = [
    ...active.filter((field) => !hiddenFields.includes(field)),
    ...(pendingField && !active.includes(pendingField) ? [pendingField] : []),
  ];
  const addable: TPickerOption[] = FILTER_FIELDS.filter(
    (item) => !unavailable.has(item.field) && !shown.includes(item.field),
  ).map((item) => {
    const Icon = fieldIcons[item.field];
    return {
      value: item.field,
      label: item.label,
      icon: <Icon className="text-muted-foreground size-4" />,
    };
  });

  return (
    <div className="scrollbar-hidden flex min-w-0 items-center gap-2 overflow-x-auto">
      {shown.map((field) => (
        <FilterChip
          key={field}
          field={field}
          filters={filters}
          options={options[field]}
          defaultOpen={field === pendingField}
          onClose={() => setPendingField(null)}
          onChange={onChange}
        />
      ))}
      {addable.length > 0 && (
        <PropertyPicker
          label="Filter by"
          options={addable}
          value={[]}
          onChange={([field]) => setPendingField(field as TFilterField)}
          trigger={
            <Button variant="ghost" size="sm" className="text-muted-foreground">
              <ListFilterIcon className="size-4" />
              Filter
            </Button>
          }
        />
      )}
      {active.length > 0 && (
        <Button
          variant="ghost"
          size="sm"
          className="text-muted-foreground"
          onClick={() =>
            onChange(active.reduce((acc, field) => clearFilterField(acc, field), filters))
          }
        >
          Clear
        </Button>
      )}
    </div>
  );
}

function FilterChip({
  field,
  filters,
  options,
  defaultOpen,
  onClose,
  onChange,
}: {
  field: TFilterField;
  filters: ListIssuesFilters;
  options: TPickerOption[];
  defaultOpen: boolean;
  onClose: () => void;
  onChange: (filters: ListIssuesFilters) => void;
}) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  const label = FILTER_FIELDS.find((item) => item.field === field)?.label ?? field;
  const Icon = fieldIcons[field];
  const values = fieldValues(filters, field);
  const setOpen = (open: boolean) => {
    setIsOpen(open);
    if (!open) onClose();
  };
  const remove = () => {
    onChange(clearFilterField(filters, field));
    onClose();
  };

  return (
    <div
      role="group"
      aria-label={`${label} filter`}
      className="bg-background flex h-7 items-center overflow-hidden rounded-md border text-sm"
    >
      <span className="text-muted-foreground flex items-center gap-1.5 px-2">
        <Icon className="size-3.5" />
        {label}
      </span>
      {field === "due_date" ? (
        <DueDateFilter filters={filters} onChange={onChange} open={isOpen} onOpenChange={setOpen} />
      ) : (
        <>
          <span className="text-muted-foreground border-l px-2 text-xs">
            {values.length > 1 ? "is any of" : "is"}
          </span>
          <PropertyPicker
            label={`Filter by ${label.toLowerCase()}`}
            multiple={field !== "subscriber"}
            options={options}
            value={values}
            open={isOpen}
            onOpenChange={setOpen}
            onChange={(next) => onChange(withFieldValues(filters, field, next))}
            trigger={
              <Button
                variant="ghost"
                size="sm"
                forceMinSize={false}
                className="h-full rounded-none border-l font-normal"
              >
                {chipValueText(values, options, pluralNouns[field])}
              </Button>
            }
          />
        </>
      )}
      <Button
        variant="ghost"
        size="sm"
        forceMinSize={false}
        aria-label={`Remove ${label.toLowerCase()} filter`}
        onClick={remove}
        className="h-full rounded-none border-l px-1.5"
      >
        <XIcon className="size-3.5" />
      </Button>
    </div>
  );
}

function chipValueText(values: string[], options: TPickerOption[], noun: string) {
  if (values.length === 0) return "any";
  if (values.length > 1) return `${values.length} ${noun}`;
  const option = options.find((item) => item.value === values[0]);
  return (
    <>
      {option?.icon}
      <span className="max-w-32 truncate">{option?.label ?? "Unknown"}</span>
    </>
  );
}

type TDueOperator = "before" | "after";

function DueDateFilter({
  filters,
  onChange,
  open,
  onOpenChange,
}: {
  filters: ListIssuesFilters;
  onChange: (filters: ListIssuesFilters) => void;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const operator: TDueOperator = filters.due_after && !filters.due_before ? "after" : "before";
  const value = (operator === "before" ? filters.due_before : filters.due_after) ?? "";
  const set = (nextOperator: TDueOperator, nextValue: string) => {
    const base = { ...filters, due_before: undefined, due_after: undefined };
    onChange(
      compactFilters(
        nextOperator === "before"
          ? { ...base, due_before: nextValue || undefined }
          : { ...base, due_after: nextValue || undefined },
      ),
    );
  };
  return (
    <>
      <PropertyPicker
        label="Due date operator"
        options={[
          { value: "before", label: "before" },
          { value: "after", label: "after" },
        ]}
        value={[operator]}
        onChange={([next]) => set(next as TDueOperator, value)}
        trigger={
          <Button
            variant="ghost"
            size="sm"
            forceMinSize={false}
            className="text-muted-foreground h-full rounded-none border-l text-xs font-normal"
          >
            {operator}
          </Button>
        }
      />
      <span className="contents" onClick={(event) => event.stopPropagation()}>
        <Popover open={open} onOpenChange={onOpenChange}>
          <PopoverTrigger
            render={
              <Button
                variant="ghost"
                size="sm"
                forceMinSize={false}
                className="h-full rounded-none border-l font-normal"
              >
                {value ? formatDueDate(value) : "pick a date"}
              </Button>
            }
          />
          <PopoverContent align="start" className="w-64 p-3">
            <Input
              type="date"
              aria-label="Due date"
              autoFocus
              defaultValue={value}
              onChange={(event) => set(operator, event.target.value)}
            />
          </PopoverContent>
        </Popover>
      </span>
    </>
  );
}
