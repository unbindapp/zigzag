import { BoxIcon } from "lucide-react";

import { EntityIcon } from "@/components/shared/entity-icon";

type TProps = {
  icon?: string | null;
  color?: string | null;
  size?: "sm" | "md" | "lg";
};

export function ProjectIcon({ icon, color, size }: TProps) {
  return <EntityIcon icon={icon} color={color} fallback={BoxIcon} size={size} />;
}
