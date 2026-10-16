import { ENTITY_ICONS } from "@/components/shared/entity-icons";
import { colorFromId, initials } from "@/lib/helpers/colors";

type TProps = {
  id: string;
  name: string;
  icon?: string | null;
  color?: string | null;
};

export function WorkspaceIcon({ id, name, icon, color }: TProps) {
  const Icon = icon ? ENTITY_ICONS[icon] : undefined;
  const isEmoji = icon && !Icon;
  return (
    <span
      aria-hidden
      className="inline-flex size-5 shrink-0 items-center justify-center rounded-sm text-xs leading-none font-semibold text-white"
      style={isEmoji ? undefined : { backgroundColor: color || colorFromId(id) }}
    >
      {Icon ? <Icon className="size-3.5" /> : isEmoji ? icon : initials(name).slice(0, 1)}
    </span>
  );
}
