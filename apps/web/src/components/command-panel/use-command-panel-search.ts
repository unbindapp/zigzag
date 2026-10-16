import type { LinkProps } from "@tanstack/react-router";
import type { ReactNode } from "react";

export type TCommandSearchItem = {
  id: string;
  label: string;
  description?: string;
  icon?: ReactNode;
  link: LinkProps;
};

export type TCommandSearchGroup = {
  heading: string;
  items: TCommandSearchItem[];
};

export type TCommandPanelSearch = {
  groups: TCommandSearchGroup[];
  isPending: boolean;
};

const emptySearch: TCommandPanelSearch = { groups: [], isPending: false };

export function useCommandPanelSearch(query: string): TCommandPanelSearch {
  void query;
  return emptySearch;
}
