import { useQuery } from "@tanstack/react-query";
import { FileTextIcon } from "lucide-react";
import { createElement } from "react";
import { useDebounceValue } from "usehooks-ts";

import type {
  TCommandPanelSearch,
  TCommandSearchGroup,
  TCommandSearchItem,
} from "@/components/command-panel/use-command-panel-search";
import { EntityIcon } from "@/components/shared/entity-icon";
import { ProjectIcon } from "@/components/shared/project-icon";
import { StateIcon } from "@/components/shared/state-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { searchQuery, type TSearchResults } from "@/lib/queries/search";

const SEARCH_DEBOUNCE_MS = 200;

export function groupSearchResults(
  results: TSearchResults | undefined,
  workspaceSlug: string,
): TCommandSearchGroup[] {
  if (!results) return [];
  const issues = results.issues.map((issue): TCommandSearchItem => ({
    id: issue.id,
    label: issue.title,
    description: issue.identifier,
    icon: issue.state
      ? createElement(StateIcon, {
          type: issue.state.type,
          color: issue.state.color,
          name: issue.state.name,
        })
      : undefined,
    link: {
      to: "/$workspace/issue/$identifier",
      params: { workspace: workspaceSlug, identifier: issue.identifier },
    },
  }));
  const projects = results.projects.map((project): TCommandSearchItem => ({
    id: project.id,
    label: project.name,
    icon: createElement(ProjectIcon, { icon: project.icon, color: project.color, size: "sm" }),
    link: {
      to: "/$workspace/project/$slug",
      params: { workspace: workspaceSlug, slug: project.slug },
    },
  }));
  const documents = results.documents.map((document): TCommandSearchItem => ({
    id: document.id,
    label: document.title,
    description: document.project?.name,
    icon: createElement(EntityIcon, { icon: document.icon, fallback: FileTextIcon, size: "sm" }),
    link: {
      to: "/$workspace/document/$slug",
      params: { workspace: workspaceSlug, slug: document.slug },
    },
  }));
  return [
    { heading: "Issues", items: issues },
    { heading: "Projects", items: projects },
    { heading: "Documents", items: documents },
  ].filter((group) => group.items.length > 0);
}

export function useSearch(query: string): TCommandPanelSearch {
  const { workspace } = useWorkspace();
  const trimmed = query.trim();
  const [debounced] = useDebounceValue(trimmed, SEARCH_DEBOUNCE_MS);
  const { data, isFetching } = useQuery(searchQuery(workspace.id, debounced));
  return {
    groups: debounced ? groupSearchResults(data?.results, workspace.slug) : [],
    isPending: Boolean(trimmed) && (isFetching || debounced !== trimmed),
  };
}
