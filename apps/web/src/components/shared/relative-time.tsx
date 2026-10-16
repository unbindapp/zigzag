import { useNow } from "@/components/providers/now-provider";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { formatDate, formatDateTime, formatRelative } from "@/lib/helpers/dates";

type TProps = {
  value: string | number | Date;
  mode?: "relative" | "date";
};

export function RelativeTime({ value, mode = "relative" }: TProps) {
  const now = useNow();
  const date = value instanceof Date ? value : new Date(value);
  const text = mode === "relative" ? formatRelative(date, now) : formatDate(date);
  return (
    <SimpleTooltip content={formatDateTime(date)}>
      <time dateTime={date.toISOString()} className="whitespace-nowrap">
        {text}
      </time>
    </SimpleTooltip>
  );
}
