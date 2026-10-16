import { useHotkey } from "@tanstack/react-hotkeys";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { createContext, type ReactNode, useContext, useMemo } from "react";

import { IssueDetail } from "@/components/issues/issue-detail";
import { SplitView } from "@/components/layout/split-view";
import { useDeviceSize } from "@/components/providers/device-size-provider";

type TIssuePeekContext = {
  peek: string | undefined;
  openPeek: (identifier: string) => void;
  closePeek: () => void;
};

const IssuePeekContext = createContext<TIssuePeekContext | null>(null);

export function useIssuePeek() {
  return useContext(IssuePeekContext);
}

export function useIsPeekable() {
  const { size } = useDeviceSize();
  return size === "lg" || size === "xl";
}

export function IssuePeekHost({ children }: { children: ReactNode }) {
  const search = useSearch({ strict: false }) as { peek?: string };
  const navigate = useNavigate();
  const isPeekable = useIsPeekable();
  const peek = search.peek;

  const value = useMemo<TIssuePeekContext>(
    () => ({
      peek,
      openPeek: (identifier) =>
        void navigate({
          to: ".",
          search: (prev: Record<string, unknown>) => ({ ...prev, peek: identifier }),
          replace: true,
        }),
      closePeek: () =>
        void navigate({
          to: ".",
          search: (prev: Record<string, unknown>) => ({ ...prev, peek: undefined }),
          replace: true,
        }),
    }),
    [peek, navigate],
  );

  return (
    <IssuePeekContext.Provider value={value}>
      <SplitView
        storageKey="issue-peek-width"
        main={children}
        aside={peek && isPeekable ? <IssuePeek identifier={peek} /> : undefined}
      />
    </IssuePeekContext.Provider>
  );
}

function IssuePeek({ identifier }: { identifier: string }) {
  const peek = useIssuePeek();
  useHotkey("Escape", () => peek?.closePeek(), {
    enabled: Boolean(peek),
    conflictBehavior: "allow",
  });
  return (
    <section
      aria-label={`Issue ${identifier}`}
      className="bg-background flex min-h-0 w-full flex-1 flex-col overflow-y-auto"
    >
      <IssueDetail identifier={identifier} variant="peek" onClose={() => peek?.closePeek()} />
    </section>
  );
}
