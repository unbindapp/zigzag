import { Skeleton } from "@/components/ui/skeleton";

const widths = ["w-1/2", "w-2/3", "w-1/3", "w-3/5", "w-2/5", "w-1/2", "w-3/4", "w-1/4"];

export function ListSkeleton({ rows = 6 }: { rows?: number }) {
  return (
    <div
      data-slot="list-skeleton"
      aria-busy
      aria-label="Loading"
      className="flex w-full flex-col overflow-hidden rounded-lg border"
    >
      {Array.from({ length: rows }, (_, index) => (
        <div key={index} className="flex h-9 items-center gap-2 border-b px-3 last:border-b-0">
          <Skeleton className="size-4 rounded-full" />
          <Skeleton className={`h-3 ${widths[index % widths.length]}`} />
        </div>
      ))}
    </div>
  );
}
