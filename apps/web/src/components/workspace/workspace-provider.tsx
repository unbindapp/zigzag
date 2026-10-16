import { createContext, ReactNode, useContext, useMemo } from "react";

import type { TMeWorkspace } from "@/lib/queries/me";
import type { TWorkspaceRole } from "@/lib/queries/workspaces";

type TWorkspaceContext = {
  workspace: TMeWorkspace;
  role: TWorkspaceRole;
  canEdit: boolean;
  isAdmin: boolean;
  isSystemAdmin: boolean;
};

const WorkspaceContext = createContext<TWorkspaceContext | null>(null);

export function WorkspaceProvider({
  workspace,
  isSystemAdmin,
  children,
}: {
  workspace: TMeWorkspace;
  isSystemAdmin: boolean;
  children: ReactNode;
}) {
  const value = useMemo<TWorkspaceContext>(() => {
    const role: TWorkspaceRole = isSystemAdmin ? "admin" : workspace.role;
    return {
      workspace,
      role,
      canEdit: role === "admin" || role === "member",
      isAdmin: role === "admin",
      isSystemAdmin,
    };
  }, [workspace, isSystemAdmin]);
  return <WorkspaceContext.Provider value={value}>{children}</WorkspaceContext.Provider>;
}

export function useWorkspace() {
  const context = useContext(WorkspaceContext);
  if (!context) throw new Error("useWorkspace must be used within a WorkspaceProvider");
  return context;
}
