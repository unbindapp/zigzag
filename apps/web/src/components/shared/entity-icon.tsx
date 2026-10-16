import type { LucideIcon } from "lucide-react";

import { ENTITY_ICONS } from "@/components/shared/entity-icons";
import { cn } from "@/components/ui/utils";

type TProps = {
  icon?: string | null;
  color?: string | null;
  fallback: LucideIcon;
  size?: "sm" | "md" | "lg";
  label?: string;
};

const sizeClass = {
  sm: "size-3.5 text-xs",
  md: "size-4 text-sm",
  lg: "size-5 text-base",
};

export function EntityIcon({ icon, color, fallback, size = "md", label }: TProps) {
  const Icon = (icon && ENTITY_ICONS[icon]) || (!icon ? fallback : null);
  const style = color ? { color } : undefined;
  if (Icon) {
    return (
      <Icon
        role={label ? "img" : undefined}
        aria-label={label}
        aria-hidden={label ? undefined : true}
        className={cn("shrink-0", sizeClass[size], !color && "text-muted-foreground")}
        style={style}
      />
    );
  }
  return (
    <span
      role={label ? "img" : undefined}
      aria-label={label}
      className={cn(
        "inline-flex shrink-0 items-center justify-center leading-none",
        sizeClass[size],
      )}
    >
      {icon}
    </span>
  );
}
