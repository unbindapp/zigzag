import { parseISO } from "date-fns";
import { CircleUserIcon } from "lucide-react";

import {
  INITIATIVE_STATUSES,
  InitiativeStatusIcon,
  initiativeStatusLabels,
} from "@/components/initiatives/initiative-status";
import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { DatePicker, PickerTrigger } from "@/components/issues/pickers";
import { PropertyPicker, type TPickerOption } from "@/components/issues/property-picker";
import { Avatar } from "@/components/shared/avatar";
import { type TUserSummary } from "@/components/shared/user-chip";
import { formatDate } from "@/lib/helpers/dates";
import type { TInitiativeStatus } from "@/lib/queries/initiatives";

const NONE = "__none__";

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
  return (
    <DatePicker
      name="Target date"
      date={date}
      format={formatTargetDate}
      onChange={onChange}
      disabled={disabled}
    />
  );
}
