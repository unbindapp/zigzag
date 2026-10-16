import { create } from "zustand";

export type TNewIssueDefaults = {
  teamId?: string;
  stateId?: string;
  projectId?: string;
  cycleId?: string;
  parentId?: string;
};

type TNewIssueStore = {
  isOpen: boolean;
  defaults: TNewIssueDefaults;
  open: (defaults?: TNewIssueDefaults) => void;
  close: () => void;
};

export const useNewIssueStore = create<TNewIssueStore>((set) => ({
  isOpen: false,
  defaults: {},
  open: (defaults = {}) => set({ isOpen: true, defaults }),
  close: () => set({ isOpen: false }),
}));
