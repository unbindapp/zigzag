import { resolveEntityIcon } from "@/components/shared/entity-icons";
import { colorFromId, initials } from "@/lib/helpers/colors";

type TProps = {
  id: string;
  name: string;
  icon?: string | null;
  color?: string | null;
  iconUrl?: string | null;
};

export function WorkspaceIcon({ id, name, icon, color, iconUrl }: TProps) {
  if (iconUrl) {
    return (
      <img aria-hidden src={iconUrl} alt="" className="size-5 shrink-0 rounded-md object-cover" />
    );
  }
  const resolved = resolveEntityIcon(icon);
  const isEmoji = resolved.kind === "emoji";
  return (
    <span
      aria-hidden
      className="inline-flex size-5 shrink-0 items-center justify-center rounded-sm text-xs leading-none font-semibold text-white"
      style={isEmoji ? undefined : { backgroundColor: color || colorFromId(id) }}
    >
      {resolved.kind === "lucide" && <resolved.Icon className="size-3.5" />}
      {resolved.kind === "emoji" && resolved.emoji}
      {resolved.kind === "none" && initials(name).slice(0, 1)}
    </span>
  );
}
