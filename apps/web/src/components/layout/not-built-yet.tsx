import { HammerIcon } from "lucide-react";

import { EmptyState } from "@/components/layout/empty-state";

export function NotBuiltYet() {
  return <EmptyState icon={HammerIcon} title="Not built yet" />;
}
