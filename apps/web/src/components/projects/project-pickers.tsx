import { useQuery } from "@tanstack/react-query";
import { CalendarIcon, CircleUserIcon, UsersIcon, XIcon } from "lucide-react";
import { useState } from "react";

import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { PropertyPicker, type TPickerOption } from "@/components/issues/property-picker";
import {
  formatProjectDate,
  RESOLUTIONS,
  resolutionLabels,
  snapToResolution,
  type TResolution,
} from "@/components/projects/project-dates";
import { ProjectStatusIcon } from "@/components/projects/project-status-icon";
import { Avatar } from "@/components/shared/avatar";
import { TeamIcon } from "@/components/shared/team-icon";
import { type TUserSummary } from "@/components/shared/user-chip";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TProject, type TProjectStatus, projectStatusesQuery } from "@/lib/queries/projects";
import { PickerTrigger } from "@/components/issues/pickers";

export type TProjectPickerAppearance = "row" | "compact";

const NONE = "__none__";

function UserIcon({ user }: { user: TUserSummary }) {
  return (
    <Avatar
      id={user.id}
      name={user.name || user.display_name || "?"}
      src={user.avatar_url}
      size="xs"
    />
  );
}

function userOption(user: TUserSummary): TPickerOption {
  return {
    value: user.id,
    label: user.name || user.display_name || "?",
    description: user.display_name ? `@${user.display_name}` : undefined,
    keywords: user.display_name ? [user.display_name] : undefined,
    icon: <UserIcon user={user} />,
  };
}

type TControl = { appearance?: TProjectPickerAppearance; disabled?: boolean };

export function useProjectStatuses() {
  const { workspace } = useWorkspace();
  const { data, isPending } = useQuery(projectStatusesQuery(workspace.id));
  return { statuses: data?.statuses ?? [], isPending };
}

export function ProjectStatusPicker({
  status,
  onChange,
  appearance = "row",
  disabled,
}: TControl & { status: TProjectStatus | undefined; onChange: (status: TProjectStatus) => void }) {
  const { statuses, isPending } = useProjectStatuses();
  const options: TPickerOption[] = statuses.map((option) => ({
    value: option.id,
    label: option.name,
    icon: <ProjectStatusIcon type={option.type} color={option.color} />,
  }));
  return (
    <PropertyPicker
      label="Change status"
      options={options}
      value={status ? [status.id] : []}
      isPending={isPending}
      disabled={disabled}
      onChange={([id]) => {
        const next = statuses.find((option) => option.id === id);
        if (next && next.id !== status?.id) onChange(next);
      }}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={status ? `Status: ${status.name}` : "Set status"}
          icon={
            status ? (
              <ProjectStatusIcon type={status.type} color={status.color} />
            ) : (
              <CalendarIcon />
            )
          }
          text={status?.name ?? "Status"}
          muted={!status}
        />
      }
    />
  );
}

export function LeadPicker({
  lead,
  onChange,
  appearance = "row",
  disabled,
}: TControl & {
  lead: TUserSummary | null | undefined;
  onChange: (user: TUserSummary | null) => void;
}) {
  const { users, isPending } = useWorkspaceUsers();
  const options: TPickerOption[] = [
    {
      value: NONE,
      label: "No lead",
      icon: <CircleUserIcon className="text-muted-foreground size-4" />,
    },
    ...users.map(userOption),
  ];
  const name = lead ? lead.name || lead.display_name : null;
  return (
    <PropertyPicker
      label="Set lead"
      options={options}
      value={[lead?.id ?? NONE]}
      isPending={isPending}
      disabled={disabled}
      onChange={([id]) => {
        if (id === (lead?.id ?? NONE)) return;
        onChange(users.find((user) => user.id === id) ?? null);
      }}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label={name ? `Lead: ${name}` : "Set lead"}
          icon={lead ? <UserIcon user={lead} /> : <CircleUserIcon className="size-4" />}
          text={name ?? "Lead"}
          muted={!lead}
        />
      }
    />
  );
}

export function MembersPicker({
  members,
  onChange,
  appearance = "row",
  disabled,
}: TControl & { members: TUserSummary[]; onChange: (ids: string[]) => void }) {
  const { users, isPending } = useWorkspaceUsers();
  return (
    <PropertyPicker
      label="Change members"
      multiple
      options={users.map(userOption)}
      value={members.map((member) => member.id)}
      isPending={isPending}
      disabled={disabled}
      onChange={onChange}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label="Change members"
          icon={
            members.length > 0 ? <AvatarStack users={members} /> : <UsersIcon className="size-4" />
          }
          text={
            members.length === 0
              ? "Members"
              : `${members.length} member${members.length === 1 ? "" : "s"}`
          }
          muted={members.length === 0}
        />
      }
    />
  );
}

export function AvatarStack({ users, max = 3 }: { users: TUserSummary[]; max?: number }) {
  const shown = users.slice(0, max);
  return (
    <span className="flex shrink-0 items-center -space-x-1">
      {shown.map((user) => (
        <UserIcon key={user.id} user={user} />
      ))}
    </span>
  );
}

export function TeamsPicker({
  teams,
  onChange,
  appearance = "row",
  disabled,
}: TControl & { teams: TProject["teams"]; onChange: (ids: string[]) => void }) {
  const { teams: available, isPending } = useTeams();
  const options: TPickerOption[] = available.map((team) => ({
    value: team.id,
    label: team.name,
    description: team.key,
    keywords: [team.key],
    icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
  }));
  return (
    <PropertyPicker
      label="Change teams"
      multiple
      options={options}
      value={teams.map((team) => team.id)}
      isPending={isPending}
      disabled={disabled}
      onChange={(ids) => {
        if (ids.length === 0) return;
        onChange(ids);
      }}
      trigger={
        <PickerTrigger
          appearance={appearance}
          label="Change teams"
          icon={
            teams[0] ? (
              <TeamIcon
                teamKey={teams[0].key}
                icon={teams[0].icon}
                color={teams[0].color}
                size="sm"
              />
            ) : (
              <UsersIcon className="size-4" />
            )
          }
          text={teams.length === 0 ? "Teams" : teams.map((team) => team.name).join(", ")}
          muted={teams.length === 0}
        />
      }
    />
  );
}

const resolutionItems = RESOLUTIONS.map((value) => ({ value, label: resolutionLabels[value] }));

export function ProjectDatePicker({
  label,
  date,
  resolution,
  onChange,
  appearance = "row",
  disabled,
}: TControl & {
  label: string;
  date: string | null;
  resolution: TResolution;
  onChange: (value: { date: string | null; resolution: TResolution }) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [draftResolution, setDraftResolution] = useState<TResolution>(resolution);
  const text = date ? formatProjectDate(date, resolution) : label;
  return (
    <Popover
      open={isOpen}
      onOpenChange={(open) => {
        setIsOpen(open);
        if (open) setDraftResolution(resolution);
      }}
    >
      <PopoverTrigger
        disabled={disabled}
        render={
          <PickerTrigger
            appearance={appearance}
            label={date ? `${label}: ${text}` : `Set ${label.toLowerCase()}`}
            icon={<CalendarIcon className="size-4" />}
            text={text}
            muted={!date}
          />
        }
      />
      <PopoverContent align="start" className="flex w-64 flex-col gap-2 p-3">
        <Select
          items={resolutionItems}
          value={draftResolution}
          onValueChange={(value) => {
            if (!value) return;
            const next = value as TResolution;
            setDraftResolution(next);
            if (date) onChange({ date: snapToResolution(date, next), resolution: next });
          }}
        >
          <SelectTrigger size="sm" aria-label={`${label} precision`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {resolutionItems.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Input
          type="date"
          aria-label={label}
          defaultValue={date ?? ""}
          onChange={(event) => {
            if (!event.target.value) return;
            onChange({
              date: snapToResolution(event.target.value, draftResolution),
              resolution: draftResolution,
            });
          }}
          onKeyDown={(event) => {
            if (event.key === "Enter") setIsOpen(false);
          }}
        />
        {date && (
          <Button
            variant="ghost"
            size="sm"
            className="self-start"
            onClick={() => {
              onChange({ date: null, resolution: "day" });
              setIsOpen(false);
            }}
          >
            <XIcon className="size-4" />
            Remove {label.toLowerCase()}
          </Button>
        )}
      </PopoverContent>
    </Popover>
  );
}
