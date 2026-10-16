import type { LinkProps } from "@tanstack/react-router";
import type { ReactNode } from "react";

import { useSearch } from "@/components/search/use-search";

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

export function useCommandPanelSearch(query: string): TCommandPanelSearch {
  return useSearch(query);
}
