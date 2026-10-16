import { useQuery } from "@tanstack/react-query";
import { parseISO } from "date-fns";
import {
  CalendarIcon,
  CircleDashedIcon,
  CircleUserIcon,
  DiamondIcon,
  type LucideIcon,
  CornerLeftUpIcon,
  TagIcon,
  TriangleIcon,
  UsersIcon,
  XIcon,
} from "lucide-react";
import { type ReactNode, useState } from "react";

import { useDebounceValue } from "usehooks-ts";

import { NONE_KEY, PRIORITY_ORDER } from "@/components/issues/grouping";
import {
  useCycles,
  useLabels,
  useMilestones,
  useProjects,
  useTeam,
  useTeamStates,
  useWorkspaceUsers,
} from "@/components/issues/picker-sources";
import { PropertyPicker, type TPickerOption } from "@/components/issues/property-picker";
import {
  type TIssue,
  type TLabelSummary,
  type TUserSummary,
  type TWorkflowStateSummary,
  searchIssuesQuery,
} from "@/lib/queries/issues";
import { Avatar } from "@/components/shared/avatar";
import { CycleIcon } from "@/components/shared/cycle-icon";
import { LabelChip } from "@/components/shared/label-chip";
import { PriorityIcon, priorityLabel } from "@/components/shared/priority-icon";
import { ProjectIcon } from "@/components/shared/project-icon";
import { StateIcon } from "@/components/shared/state-icon";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { formatDate } from "@/lib/helpers/dates";
import { estimateOptions, formatEstimate, type TEstimateType } from "@/lib/helpers/estimates";
import type { TLabel, TLabelTree } from "@/lib/queries/labels";

export type TPickerAppearance = "row" | "compact" | "icon";

type TPickerControl = {
  appearance?: TPickerAppearance;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  disabled?: boolean;
};

function PickerTrigger({
  appearance,
  label,
  icon,
  text,
  muted,
  ...props
}: {
  appearance: TPickerAppearance;
  label: string;
  icon: ReactNode;
  text: ReactNode;
  muted?: boolean;
}) {
  switch (appearance) {
    case "icon":
      return (
        <Button
          variant="ghost"
          size="icon"
          forceMinSize={false}
          aria-label={label}
          className="shrink-0"
          {...props}
        >
          {icon}
        </Button>
      );
    case "compact":
      return (
        <Button
          variant="outline"
          size="sm"
          forceMinSize={false}
          aria-label={label}
          data-muted={muted || undefined}
          className="data-muted:text-muted-foreground max-w-48 font-normal"
          {...props}
        >
          {icon}
          <span className="truncate">{text}</span>
        </Button>
      );
    case "row":
      return (
        <Button
          variant="ghost"
          size="sm"
          forceMinSize={false}
          aria-label={label}
          data-muted={muted || undefined}
          className="data-muted:text-muted-foreground w-full min-w-0 justify-start font-normal"
          {...props}
        >
          {icon}
          <span className="min-w-0 truncate">{text}</span>
        </Button>
      );
  }
}

function MutedIcon({ icon: Icon }: { icon: LucideIcon }) {
  return <Icon className="text-muted-foreground size-4" />;
}

export function StatePicker({
  teamId,
  state,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  teamId: string;
  state: TWorkflowStateSummary | undefined;
  onChange: (state: TWorkflowStateSummary) => void;
}) {
  const states = useTeamStates(teamId);
  const options: TPickerOption[] = states.map((item, index) => ({
    value: item.id,
    label: item.name,
    icon: <StateIcon type={item.type} color={item.color} name={item.name} />,
    shortcut: index < 9 ? String(index + 1) : undefined,
  }));
  return (
    <PropertyPicker
      label="Change status"
      options={options}
      value={state ? [state.id] : []}
      onChange={([id]) => {
        const next = states.find((item) => item.id === id);
        if (next && next.id !== state?.id) onChange(next);
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={state ? `Status: ${state.name}` : "Set status"}
          icon={
            state ? (
              <StateIcon type={state.type} color={state.color} name={state.name} />
            ) : (
              <MutedIcon icon={CircleDashedIcon} />
            )
          }
          text={state?.name ?? "Status"}
          muted={!state}
        />
      }
    />
  );
}

export function PriorityPicker({
  priority,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & { priority: number; onChange: (priority: number) => void }) {
  const options: TPickerOption[] = [0, ...PRIORITY_ORDER.filter((p) => p !== 0)].map((value) => ({
    value: String(value),
    label: priorityLabel(value),
    icon: <PriorityIcon priority={value} />,
    shortcut: String(value),
  }));
  return (
    <PropertyPicker
      label="Change priority"
      options={options}
      value={[String(priority)]}
      onChange={([value]) => {
        const next = Number(value);
        if (next !== priority) onChange(next);
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={`Priority: ${priorityLabel(priority)}`}
          icon={<PriorityIcon priority={priority} />}
          text={priority === 0 ? "Priority" : priorityLabel(priority)}
          muted={priority === 0}
        />
      }
    />
  );
}

function userOption(user: TUserSummary): TPickerOption {
  return {
    value: user.id,
    label: user.name || user.display_name,
    description: user.display_name ? `@${user.display_name}` : undefined,
    keywords: [user.display_name],
    icon: (
      <Avatar id={user.id} name={user.name || user.display_name} src={user.avatar_url} size="xs" />
    ),
  };
}

export function UserAvatar({
  user,
  size = "sm",
}: {
  user: TUserSummary;
  size?: "xs" | "sm" | "md";
}) {
  return (
    <Avatar id={user.id} name={user.name || user.display_name} src={user.avatar_url} size={size} />
  );
}

export function AssigneePicker({
  assignee,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  assignee: TUserSummary | null | undefined;
  onChange: (user: TUserSummary | null) => void;
}) {
  const { users, isPending } = useWorkspaceUsers();
  const options: TPickerOption[] = [
    {
      value: NONE_KEY,
      label: "No assignee",
      icon: <MutedIcon icon={CircleUserIcon} />,
    },
    ...users.map(userOption),
  ];
  const name = assignee ? assignee.name || assignee.display_name : null;
  return (
    <PropertyPicker
      label="Assign to"
      options={options}
      value={[assignee?.id ?? NONE_KEY]}
      isPending={isPending}
      onChange={([id]) => {
        if (id === (assignee?.id ?? NONE_KEY)) return;
        onChange(users.find((user) => user.id === id) ?? null);
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={name ? `Assignee: ${name}` : "Assign"}
          icon={assignee ? <UserAvatar user={assignee} /> : <MutedIcon icon={CircleUserIcon} />}
          text={name ?? "Assignee"}
          muted={!assignee}
        />
      }
    />
  );
}

export function SubscribersPicker({
  subscribers,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  subscribers: TUserSummary[];
  onChange: (added: TUserSummary[], removed: TUserSummary[]) => void;
}) {
  const { users, isPending } = useWorkspaceUsers();
  const currentIds = subscribers.map((user) => user.id);
  return (
    <PropertyPicker
      label="Subscribers"
      multiple
      options={users.map(userOption)}
      value={currentIds}
      isPending={isPending}
      onChange={(ids) => {
        const added = users.filter(
          (user) => ids.includes(user.id) && !currentIds.includes(user.id),
        );
        const removed = subscribers.filter((user) => !ids.includes(user.id));
        onChange(added, removed);
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label="Change subscribers"
          icon={
            subscribers.length > 0 ? (
              <span className="flex items-center -space-x-1">
                {subscribers.slice(0, 3).map((user) => (
                  <UserAvatar key={user.id} user={user} size="xs" />
                ))}
              </span>
            ) : (
              <MutedIcon icon={UsersIcon} />
            )
          }
          text={
            subscribers.length === 0
              ? "No subscribers"
              : `${subscribers.length} subscriber${subscribers.length === 1 ? "" : "s"}`
          }
          muted={subscribers.length === 0}
        />
      }
    />
  );
}

type TFlatLabel = Pick<
  TLabel,
  "id" | "name" | "color" | "parent_id" | "archived_at" | "is_group" | "group_single_select"
>;

export function flattenLabels(labels: TLabelTree[]) {
  const groupsById = new Map<string, TFlatLabel>();
  const flat: { label: TFlatLabel; group: TFlatLabel | undefined }[] = [];
  for (const label of labels) {
    if (label.is_group) groupsById.set(label.id, label);
  }
  for (const label of labels) {
    if (label.archived_at) continue;
    if (label.is_group) {
      for (const child of label.children ?? []) {
        if (!child.archived_at) flat.push({ label: child, group: label });
      }
      continue;
    }
    flat.push({ label, group: label.parent_id ? groupsById.get(label.parent_id) : undefined });
  }
  return flat;
}

function applySingleSelectGroups(
  previous: string[],
  next: string[],
  flat: ReturnType<typeof flattenLabels>,
) {
  const added = next.filter((id) => !previous.includes(id));
  return added.reduce((ids, addedId) => {
    const entry = flat.find((item) => item.label.id === addedId);
    if (!entry?.group?.group_single_select) return ids;
    const siblings = flat
      .filter((item) => item.group?.id === entry.group?.id && item.label.id !== addedId)
      .map((item) => item.label.id);
    return ids.filter((id) => !siblings.includes(id));
  }, next);
}

export function LabelsPicker({
  teamId,
  labels,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  teamId: string | undefined;
  labels: TLabelSummary[];
  onChange: (labels: TLabelSummary[]) => void;
}) {
  const { labels: available, isPending } = useLabels(teamId);
  const flat = flattenLabels(available);
  const currentIds = labels.map((label) => label.id);
  const options: TPickerOption[] = flat.map(({ label, group }) => ({
    value: label.id,
    label: label.name,
    group: group?.name,
    keywords: group ? [group.name] : undefined,
    icon: (
      <span
        aria-hidden
        className="bg-muted-foreground size-2 shrink-0 rounded-full"
        style={{ backgroundColor: label.color }}
      />
    ),
  }));
  const text =
    labels.length === 0 ? (
      "Labels"
    ) : appearance === "row" ? (
      <span className="flex min-w-0 flex-wrap items-center gap-1.5">
        {labels.map((label) => (
          <LabelChip key={label.id} name={label.name} color={label.color} />
        ))}
      </span>
    ) : labels.length === 1 ? (
      labels[0].name
    ) : (
      `${labels.length} labels`
    );
  return (
    <PropertyPicker
      label="Change labels"
      multiple
      options={options}
      value={currentIds}
      isPending={isPending}
      emptyText="No labels"
      onChange={(ids) => {
        const resolved = applySingleSelectGroups(currentIds, ids, flat);
        onChange(
          resolved
            .map(
              (id) =>
                flat.find((item) => item.label.id === id)?.label ?? labels.find((l) => l.id === id),
            )
            .filter((label): label is TFlatLabel | TLabelSummary => Boolean(label))
            .map((label) => ({
              id: label.id,
              name: label.name,
              color: label.color,
              parent_id: label.parent_id,
            })),
        );
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label="Change labels"
          icon={<MutedIcon icon={TagIcon} />}
          text={text}
          muted={labels.length === 0}
        />
      }
    />
  );
}

type TProjectSummary = NonNullable<TIssue["project"]>;

export function ProjectPicker({
  teamId,
  project,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  teamId: string | undefined;
  project: TProjectSummary | null | undefined;
  onChange: (project: TProjectSummary | null) => void;
}) {
  const { projects, isPending } = useProjects(teamId);
  const options: TPickerOption[] = [
    { value: NONE_KEY, label: "No project", icon: <ProjectIcon size="md" /> },
    ...projects.map((item) => ({
      value: item.id,
      label: item.name,
      icon: <ProjectIcon icon={item.icon} color={item.color} />,
    })),
  ];
  return (
    <PropertyPicker
      label="Move to project"
      options={options}
      value={[project?.id ?? NONE_KEY]}
      isPending={isPending}
      onChange={([id]) => {
        if (id === (project?.id ?? NONE_KEY)) return;
        const next = projects.find((item) => item.id === id);
        onChange(
          next
            ? { id: next.id, name: next.name, slug: next.slug, icon: next.icon, color: next.color }
            : null,
        );
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={project ? `Project: ${project.name}` : "Add to project"}
          icon={<ProjectIcon icon={project?.icon} color={project?.color} />}
          text={project?.name ?? "Project"}
          muted={!project}
        />
      }
    />
  );
}

type TMilestoneSummary = NonNullable<TIssue["milestone"]>;

export function MilestonePicker({
  projectId,
  milestone,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  projectId: string | undefined;
  milestone: TMilestoneSummary | null | undefined;
  onChange: (milestone: TMilestoneSummary | null) => void;
}) {
  const { milestones, isPending } = useMilestones(projectId);
  const options: TPickerOption[] = [
    { value: NONE_KEY, label: "No milestone", icon: <MutedIcon icon={DiamondIcon} /> },
    ...milestones.map((item) => ({
      value: item.id,
      label: item.name,
      description: item.target_date ? formatDate(parseISO(item.target_date)) : undefined,
      icon: <DiamondIcon className="size-4" />,
    })),
  ];
  return (
    <PropertyPicker
      label="Set milestone"
      options={options}
      value={[milestone?.id ?? NONE_KEY]}
      isPending={isPending}
      disabled={!projectId || control.disabled}
      emptyText="This project has no milestones"
      onChange={([id]) => {
        if (id === (milestone?.id ?? NONE_KEY)) return;
        const next = milestones.find((item) => item.id === id);
        onChange(next ? { id: next.id, name: next.name } : null);
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={milestone ? `Milestone: ${milestone.name}` : "Set milestone"}
          icon={<MutedIcon icon={DiamondIcon} />}
          text={milestone?.name ?? "Milestone"}
          muted={!milestone}
        />
      }
    />
  );
}

type TCycleSummary = NonNullable<TIssue["cycle"]>;

export function cycleName(cycle: { number: number; name: string }) {
  return cycle.name || `Cycle ${cycle.number}`;
}

export function CyclePicker({
  teamId,
  cycle,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  teamId: string | undefined;
  cycle: TCycleSummary | null | undefined;
  onChange: (cycle: TCycleSummary | null) => void;
}) {
  const { cycles, isPending, enabled } = useCycles(teamId);
  if (!enabled) return null;
  const options: TPickerOption[] = [
    { value: NONE_KEY, label: "No cycle", icon: <CycleIcon /> },
    ...cycles.map((item) => ({
      value: item.id,
      label: cycleName(item),
      description: item.is_active ? "Current" : formatDate(item.starts_at),
      icon: <CycleIcon status={item.is_active ? "active" : "upcoming"} />,
    })),
  ];
  return (
    <PropertyPicker
      label="Move to cycle"
      options={options}
      value={[cycle?.id ?? NONE_KEY]}
      isPending={isPending}
      onChange={([id]) => {
        if (id === (cycle?.id ?? NONE_KEY)) return;
        const next = cycles.find((item) => item.id === id);
        onChange(next ? { id: next.id, number: next.number, name: next.name } : null);
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={cycle ? `Cycle: ${cycleName(cycle)}` : "Add to cycle"}
          icon={<CycleIcon status={cycle ? "active" : "upcoming"} />}
          text={cycle ? cycleName(cycle) : "Cycle"}
          muted={!cycle}
        />
      }
    />
  );
}

export function EstimatePicker({
  teamId,
  estimate,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  teamId: string | undefined;
  estimate: number | null;
  onChange: (estimate: number | null) => void;
}) {
  const team = useTeam(teamId);
  const type = (team?.estimate_type ?? "none") as TEstimateType;
  if (!team || type === "none") return null;
  const values = estimateOptions({
    type,
    allowZero: team.estimate_allow_zero,
    extended: team.estimate_extended,
  });
  const options: TPickerOption[] = [
    { value: NONE_KEY, label: "No estimate", icon: <MutedIcon icon={TriangleIcon} /> },
    ...values.map((value) => ({
      value: String(value),
      label:
        `${formatEstimate(value, type)} ${type === "tshirt" ? "" : value === 1 ? "point" : "points"}`.trim(),
      icon: <TriangleIcon className="size-4" />,
    })),
  ];
  const text = estimate === null ? "Estimate" : formatEstimate(estimate, type);
  return (
    <PropertyPicker
      label="Set estimate"
      options={options}
      value={[estimate === null ? NONE_KEY : String(estimate)]}
      onChange={([value]) => onChange(value === NONE_KEY ? null : Number(value))}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={estimate === null ? "Set estimate" : `Estimate: ${text}`}
          icon={<MutedIcon icon={TriangleIcon} />}
          text={text}
          muted={estimate === null}
        />
      }
    />
  );
}

export function formatDueDate(value: string) {
  return formatDate(parseISO(value));
}

export function DueDatePicker({
  dueDate,
  onChange,
  appearance = "row",
  open,
  onOpenChange,
  disabled,
}: TPickerControl & { dueDate: string | null; onChange: (dueDate: string | null) => void }) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = open ?? internalOpen;
  const setOpen = (next: boolean) => {
    setInternalOpen(next);
    onOpenChange?.(next);
  };
  const text = dueDate ? formatDueDate(dueDate) : "Due date";
  return (
    <span className="contents" onClick={(event) => event.stopPropagation()}>
      <Popover open={isOpen} onOpenChange={setOpen}>
        <PopoverTrigger
          disabled={disabled}
          render={
            <PickerTrigger
              appearance={appearance}
              label={dueDate ? `Due date: ${text}` : "Set due date"}
              icon={<MutedIcon icon={CalendarIcon} />}
              text={text}
              muted={!dueDate}
            />
          }
        />
        <PopoverContent align="start" className="flex w-64 flex-col gap-2 p-3">
          <Input
            type="date"
            aria-label="Due date"
            autoFocus
            defaultValue={dueDate ?? ""}
            onChange={(event) => {
              if (!event.target.value) return;
              onChange(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") setOpen(false);
            }}
          />
          {dueDate && (
            <Button
              variant="ghost"
              size="sm"
              className="self-start"
              onClick={() => {
                onChange(null);
                setOpen(false);
              }}
            >
              <XIcon className="size-4" />
              Remove due date
            </Button>
          )}
        </PopoverContent>
      </Popover>
    </span>
  );
}

type TParentSummary = NonNullable<TIssue["parent"]>;

export function ParentPicker({
  issueId,
  parent,
  onChange,
  appearance = "row",
  ...control
}: TPickerControl & {
  issueId?: string;
  parent: TParentSummary | null | undefined;
  onChange: (parent: TParentSummary | null) => void;
}) {
  const { workspace } = useWorkspace();
  const [query, setQuery] = useState("");
  const [debounced] = useDebounceValue(query, 200);
  const { data, isFetching } = useQuery(searchIssuesQuery(workspace.id, debounced));
  const results = (data?.issues ?? []).filter((issue) => issue.id !== issueId);
  const options: TPickerOption[] = [
    ...(parent
      ? [{ value: NONE_KEY, label: "Remove parent", icon: <MutedIcon icon={XIcon} /> }]
      : []),
    ...results.map((issue) => ({
      value: issue.id,
      label: issue.title,
      description: issue.identifier,
      icon: <StateIcon type={issue.state.type} color={issue.state.color} name={issue.state.name} />,
    })),
  ];
  return (
    <PropertyPicker
      label="Set parent issue"
      searchPlaceholder="Search issues…"
      options={options}
      value={parent ? [parent.id] : []}
      onSearchChange={setQuery}
      isPending={isFetching}
      emptyText={debounced ? "No matching issues" : "Type to search issues"}
      onChange={([id]) => {
        if (id === NONE_KEY) {
          onChange(null);
          return;
        }
        const next = results.find((issue) => issue.id === id);
        if (next) onChange({ id: next.id, identifier: next.identifier, title: next.title });
      }}
      {...control}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={parent ? `Parent: ${parent.identifier}` : "Set parent issue"}
          icon={<MutedIcon icon={CornerLeftUpIcon} />}
          text={parent ? `${parent.identifier} ${parent.title}` : "Parent issue"}
          muted={!parent}
        />
      }
    />
  );
}
