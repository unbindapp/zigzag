import { create } from "zustand";

export type TNewProjectDefaults = {
  statusId?: string;
  teamId?: string;
  initiativeId?: string;
};

type TNewProjectStore = {
  isOpen: boolean;
  defaults: TNewProjectDefaults;
  open: (defaults?: TNewProjectDefaults) => void;
  close: () => void;
};

export const useNewProjectStore = create<TNewProjectStore>((set) => ({
  isOpen: false,
  defaults: {},
  open: (defaults = {}) => set({ isOpen: true, defaults }),
  close: () => set({ isOpen: false }),
}));
