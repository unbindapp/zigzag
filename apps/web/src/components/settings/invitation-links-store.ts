import { create } from "zustand";

type TInvitationLinksStore = {
  links: Record<string, string>;
  remember: (invitationId: string, link: string) => void;
};

export const useInvitationLinksStore = create<TInvitationLinksStore>((set) => ({
  links: {},
  remember: (invitationId, link) =>
    set((state) => ({ links: { ...state.links, [invitationId]: link } })),
}));
