import { ReactNode } from "react";

import { cn } from "@/components/ui/utils";

type TProps = {
  width?: "full" | "narrow";
  children: ReactNode;
};

export function PageContainer({ width = "full", children }: TProps) {
  return (
    <div data-slot="page-container" className="flex min-h-0 w-full flex-1 flex-col overflow-y-auto">
      <div
        className={cn(
          "mx-auto flex w-full flex-1 flex-col gap-6 px-4 py-4 sm:px-6 sm:py-5",
          width === "narrow" ? "max-w-3xl" : "max-w-none",
        )}
      >
        {children}
      </div>
    </div>
  );
}
