import { TriangleIcon } from "lucide-react";

import { formatEstimate, type TEstimateType } from "@/lib/helpers/estimates";

type TProps = {
  value: number | null | undefined;
  type: TEstimateType;
};

export function EstimateBadge({ value, type }: TProps) {
  if (value === null || value === undefined || type === "none") return null;
  const text = formatEstimate(value, type);
  return (
    <span
      aria-label={`Estimate ${text}`}
      className="text-muted-foreground inline-flex h-5 shrink-0 items-center gap-1.5 rounded-full border px-2 text-xs leading-none"
    >
      <TriangleIcon className="size-3.5" />
      {text}
    </span>
  );
}
