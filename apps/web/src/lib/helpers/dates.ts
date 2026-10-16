import { format } from "date-fns";

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;
const MONTH = 30 * DAY;
const YEAR = 365 * DAY;

type TDateInput = string | number | Date;

function toDate(value: TDateInput) {
  return value instanceof Date ? value : new Date(value);
}

function compactDistance(ms: number) {
  if (ms < HOUR) return `${Math.floor(ms / MINUTE)}m`;
  if (ms < DAY) return `${Math.floor(ms / HOUR)}h`;
  if (ms < WEEK) return `${Math.floor(ms / DAY)}d`;
  if (ms < MONTH) return `${Math.floor(ms / WEEK)}w`;
  if (ms < YEAR) return `${Math.floor(ms / MONTH)}mo`;
  return `${Math.floor(ms / YEAR)}y`;
}

export function formatRelative(value: TDateInput, now: TDateInput = Date.now()) {
  const diff = toDate(now).getTime() - toDate(value).getTime();
  if (Math.abs(diff) < MINUTE) return "just now";
  if (diff > 0) return `${compactDistance(diff)} ago`;
  return `in ${compactDistance(-diff)}`;
}

export function formatDate(value: TDateInput) {
  return format(toDate(value), "MMM d, yyyy");
}

export function formatDateTime(value: TDateInput) {
  return format(toDate(value), "MMM d, yyyy, h:mm a");
}
