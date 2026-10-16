import { type LinkProps } from "@tanstack/react-router";
import {
  BoxIcon,
  CircleDotIcon,
  FileTextIcon,
  LayersIcon,
  type LucideIcon,
  RefreshCwIcon,
  TagIcon,
  TargetIcon,
  UsersIcon,
} from "lucide-react";

import type { TFavorite, TFavoriteKind } from "@/lib/queries/favorites";

export const favoriteKindIcons: Record<TFavoriteKind, LucideIcon> = {
  issue: CircleDotIcon,
  project: BoxIcon,
  cycle: RefreshCwIcon,
  view: LayersIcon,
  document: FileTextIcon,
  label: TagIcon,
  team: UsersIcon,
  initiative: TargetIcon,
};

export function favoriteLink(workspace: string, favorite: TFavorite): LinkProps | null {
  const params = favorite.url_params;
  switch (favorite.kind) {
    case "issue":
      return {
        to: "/$workspace/issue/$identifier",
        params: { workspace, identifier: params.identifier },
      };
    case "project":
      return { to: "/$workspace/project/$slug", params: { workspace, slug: params.slug } };
    case "cycle":
      return {
        to: "/$workspace/team/$key/cycle/$number",
        params: { workspace, key: params.key, number: params.number },
      };
    case "view":
      return { to: "/$workspace/view/$view_id", params: { workspace, view_id: params.view_id } };
    case "document":
      return { to: "/$workspace/document/$slug", params: { workspace, slug: params.slug } };
    case "label":
      return {
        to: "/$workspace/issues",
        params: { workspace },
        search: { label_ids: [params.label_id] },
      };
    case "team":
      return { to: "/$workspace/team/$key/issues", params: { workspace, key: params.key } };
    case "initiative":
      return { to: "/$workspace/initiative/$slug", params: { workspace, slug: params.slug } };
  }
  return null;
}
