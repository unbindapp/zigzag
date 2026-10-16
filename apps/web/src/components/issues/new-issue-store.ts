import { useEffect } from "react";
import { create } from "zustand";

import type { TIssueDraft } from "@/components/issues/issue-draft";

export type TNewIssueDefaults = {
  teamId?: string;
  stateId?: string;
  projectId?: string;
  cycleId?: string;
  parentId?: string;
  draft?: Partial<TIssueDraft>;
};

type TNewIssueStore = {
  isOpen: boolean;
  defaults: TNewIssueDefaults;
  context: TNewIssueDefaults;
  open: (defaults?: TNewIssueDefaults) => void;
  close: () => void;
  setContext: (context: TNewIssueDefaults) => void;
};

export const useNewIssueStore = create<TNewIssueStore>((set, get) => ({
  isOpen: false,
  defaults: {},
  context: {},
  open: (defaults = {}) => {
    const { context } = get();
    set({
      isOpen: true,
      defaults: { ...context, ...defaults, draft: { ...context.draft, ...defaults.draft } },
    });
  },
  close: () => set({ isOpen: false }),
  setContext: (context) => set({ context }),
}));

export function useNewIssueContext(context: TNewIssueDefaults) {
  const setContext = useNewIssueStore((state) => state.setContext);
  const serialized = JSON.stringify(context);
  useEffect(() => {
    setContext(JSON.parse(serialized) as TNewIssueDefaults);
    return () => setContext({});
  }, [serialized, setContext]);
}
