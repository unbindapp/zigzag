import { LucideIcon } from "lucide-react";
import { ReactNode } from "react";

import { EmptyState } from "@/components/layout/empty-state";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";

type TProps = {
  isPending: boolean;
  error: Error | null;
  isEmpty: boolean;
  onRetry?: () => void;
  empty: {
    icon: LucideIcon;
    title: string;
    description?: string;
    action?: ReactNode;
  } | null;
  skeletonRows?: number;
  children: ReactNode;
};

export function ListStates({
  isPending,
  error,
  isEmpty,
  onRetry,
  empty,
  skeletonRows,
  children,
}: TProps) {
  if (isPending) return <ListSkeleton rows={skeletonRows} />;
  if (error) return <ErrorCard message={error.message} onRetry={onRetry} />;
  if (isEmpty) return empty && <EmptyState {...empty} />;
  return children;
}
