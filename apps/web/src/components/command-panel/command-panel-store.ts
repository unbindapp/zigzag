import { create } from "zustand";

type TCommandPanelStore = {
  isOpen: boolean;
  setOpen: (isOpen: boolean) => void;
};

export const useCommandPanelStore = create<TCommandPanelStore>((set) => ({
  isOpen: false,
  setOpen: (isOpen) => set({ isOpen }),
}));
