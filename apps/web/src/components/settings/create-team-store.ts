import { create } from "zustand";

type TCreateTeamStore = {
  isOpen: boolean;
  setOpen: (isOpen: boolean) => void;
};

export const useCreateTeamStore = create<TCreateTeamStore>((set) => ({
  isOpen: false,
  setOpen: (isOpen) => set({ isOpen }),
}));
