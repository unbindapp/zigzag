import { useEffect } from "react";

const key = "last_workspace_slug";

export function getLastWorkspaceSlug(): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function useRememberWorkspace(slug: string) {
  useEffect(() => {
    try {
      localStorage.setItem(key, slug);
    } catch {
      return;
    }
  }, [slug]);
}
