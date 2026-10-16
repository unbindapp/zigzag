import { create } from "zustand";

type TIssueSelectionStore = {
  selectedIds: string[];
  anchorId: string | null;
  focusedId: string | null;
  toggle: (id: string) => void;
  selectRange: (orderedIds: string[], id: string) => void;
  setSelected: (ids: string[]) => void;
  setFocused: (id: string | null) => void;
  clear: () => void;
};

export const useIssueSelectionStore = create<TIssueSelectionStore>((set, get) => ({
  selectedIds: [],
  anchorId: null,
  focusedId: null,
  toggle: (id) => {
    const { selectedIds } = get();
    const isSelected = selectedIds.includes(id);
    set({
      selectedIds: isSelected ? selectedIds.filter((item) => item !== id) : [...selectedIds, id],
      anchorId: id,
    });
  },
  selectRange: (orderedIds, id) => {
    const { anchorId, selectedIds } = get();
    const from = anchorId ? orderedIds.indexOf(anchorId) : -1;
    const to = orderedIds.indexOf(id);
    if (from === -1 || to === -1) {
      get().toggle(id);
      return;
    }
    const range = orderedIds.slice(Math.min(from, to), Math.max(from, to) + 1);
    set({ selectedIds: [...new Set([...selectedIds, ...range])] });
  },
  setSelected: (ids) => set({ selectedIds: ids }),
  setFocused: (id) => set({ focusedId: id }),
  clear: () => set({ selectedIds: [], anchorId: null }),
}));
