import { ENTITY_ICONS } from "@/components/shared/entity-icons";
import { cn } from "@/components/ui/utils";
import { colorFromId } from "@/lib/helpers/colors";

type TProps = {
  teamKey: string;
  icon?: string | null;
  color?: string | null;
  size?: "sm" | "md";
};

export function TeamIcon({ teamKey, icon, color, size = "md" }: TProps) {
  const resolvedColor = color || colorFromId(teamKey);
  const Icon = icon ? ENTITY_ICONS[icon] : undefined;
  return (
    <span
      aria-hidden
      data-size={size}
      className={cn(
        "inline-flex size-5 shrink-0 items-center justify-center rounded-sm text-xs leading-none font-semibold text-white data-[size=sm]:size-4",
        icon && !Icon && "bg-transparent",
      )}
      style={icon && !Icon ? undefined : { backgroundColor: resolvedColor }}
    >
      {Icon ? <Icon className="size-3.5" /> : icon ? icon : teamKey.slice(0, 1)}
    </span>
  );
}
