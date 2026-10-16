import { parseISO } from "date-fns";
import { CalendarIcon, CircleUserIcon, XIcon } from "lucide-react";
import { type ReactNode, useState } from "react";

import {
  INITIATIVE_STATUSES,
  InitiativeStatusIcon,
  initiativeStatusLabels,
} from "@/components/initiatives/initiative-status";
import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { PropertyPicker, type TPickerOption } from "@/components/issues/property-picker";
import { Avatar } from "@/components/shared/avatar";
import { type TUserSummary } from "@/components/shared/user-chip";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { formatDate } from "@/lib/helpers/dates";
import type { TInitiativeStatus } from "@/lib/queries/initiatives";

const NONE = "__none__";

type TTriggerProps = {
  label: string;
  icon: ReactNode;
  text: ReactNode;
  muted?: boolean;
};

function PickerTrigger({ label, icon, text, muted, ...props }: TTriggerProps) {
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

export function userName(user: TUserSummary) {
  return user.name || user.display_name || "Unknown";
}

export function formatTargetDate(value: string) {
  return formatDate(parseISO(value));
}

export function InitiativeStatusPicker({
  status,
  onChange,
  disabled,
}: {
  status: TInitiativeStatus;
  onChange: (status: TInitiativeStatus) => void;
  disabled?: boolean;
}) {
  const options: TPickerOption[] = INITIATIVE_STATUSES.map((option) => ({
    value: option,
    label: initiativeStatusLabels[option],
    icon: <InitiativeStatusIcon status={option} />,
  }));
  return (
    <PropertyPicker
      label="Change status"
      options={options}
      value={[status]}
      disabled={disabled}
      onChange={([next]) => {
        if (!next || next === status) return;
        onChange(next as TInitiativeStatus);
      }}
      trigger={
        <PickerTrigger
          label={`Status: ${initiativeStatusLabels[status]}`}
          icon={<InitiativeStatusIcon status={status} />}
          text={initiativeStatusLabels[status]}
        />
      }
    />
  );
}

export function OwnerPicker({
  owner,
  onChange,
  disabled,
}: {
  owner: TUserSummary | null | undefined;
  onChange: (user: TUserSummary | null) => void;
  disabled?: boolean;
}) {
  const { users, isPending } = useWorkspaceUsers();
  const options: TPickerOption[] = [
    {
      value: NONE,
      label: "No owner",
      icon: <CircleUserIcon className="text-muted-foreground size-4" />,
    },
    ...users.map((user) => ({
      value: user.id,
      label: userName(user),
      description: user.display_name ? `@${user.display_name}` : undefined,
      keywords: user.display_name ? [user.display_name] : undefined,
      icon: <Avatar id={user.id} name={userName(user)} src={user.avatar_url} size="xs" />,
    })),
  ];
  return (
    <PropertyPicker
      label="Set owner"
      options={options}
      value={[owner?.id ?? NONE]}
      isPending={isPending}
      disabled={disabled}
      onChange={([id]) => {
        if (id === (owner?.id ?? NONE)) return;
        onChange(users.find((user) => user.id === id) ?? null);
      }}
      trigger={
        <PickerTrigger
          label={owner ? `Owner: ${userName(owner)}` : "Set owner"}
          icon={
            owner ? (
              <Avatar id={owner.id} name={userName(owner)} src={owner.avatar_url} size="xs" />
            ) : (
              <CircleUserIcon className="size-4" />
            )
          }
          text={owner ? userName(owner) : "Owner"}
          muted={!owner}
        />
      }
    />
  );
}

export function TargetDatePicker({
  date,
  onChange,
  disabled,
}: {
  date: string | null;
  onChange: (date: string | null) => void;
  disabled?: boolean;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const text = date ? formatTargetDate(date) : "Target date";
  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger
        disabled={disabled}
        render={
          <PickerTrigger
            label={date ? `Target date: ${text}` : "Set target date"}
            icon={<CalendarIcon className="size-4" />}
            text={text}
            muted={!date}
          />
        }
      />
      <PopoverContent align="start" className="flex w-64 flex-col gap-2 p-3">
        <Input
          type="date"
          aria-label="Target date"
          autoFocus
          defaultValue={date ?? ""}
          onChange={(event) => {
            if (!event.target.value) return;
            onChange(event.target.value);
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
              onChange(null);
              setIsOpen(false);
            }}
          >
            <XIcon className="size-4" />
            Remove target date
          </Button>
        )}
      </PopoverContent>
    </Popover>
  );
}
