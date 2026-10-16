import { addDays, addHours, nextMonday, set } from "date-fns";

export type TSnoozeOption = {
  id: "hour" | "tomorrow" | "next-week";
  label: string;
  until: Date;
};

const MORNING = { hours: 9, minutes: 0, seconds: 0, milliseconds: 0 };

export function snoozeOptions(now: Date): TSnoozeOption[] {
  return [
    { id: "hour", label: "In 1 hour", until: addHours(now, 1) },
    { id: "tomorrow", label: "Tomorrow, 9:00", until: set(addDays(now, 1), MORNING) },
    { id: "next-week", label: "Next week, Monday 9:00", until: set(nextMonday(now), MORNING) },
  ];
}
