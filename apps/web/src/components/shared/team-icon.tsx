import { resolveEntityIcon } from "@/components/shared/entity-icons";
import { colorFromId } from "@/lib/helpers/colors";

type TProps = {
  teamKey: string;
  icon?: string | null;
  color?: string | null;
  size?: "sm" | "md";
};

export function TeamIcon({ teamKey, icon, color, size = "md" }: TProps) {
  const resolved = resolveEntityIcon(icon);
  const isEmoji = resolved.kind === "emoji";
  return (
    <span
      aria-hidden
      data-size={size}
      className="inline-flex size-5 shrink-0 items-center justify-center rounded-sm text-xs leading-none font-semibold text-white data-[size=sm]:size-4"
      style={isEmoji ? undefined : { backgroundColor: color || colorFromId(teamKey) }}
    >
      {resolved.kind === "lucide" && <resolved.Icon className="size-3.5" />}
      {resolved.kind === "emoji" && resolved.emoji}
      {resolved.kind === "none" && teamKey.slice(0, 1)}
    </span>
  );
}
