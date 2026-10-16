import { MenuIcon, SearchIcon } from "lucide-react";
import { ReactNode, useState } from "react";

import { CommandPanel } from "@/components/command-panel/command-panel";
import { useCommandPanelStore } from "@/components/command-panel/command-panel-store";
import { NewIssueDialog } from "@/components/issues/new-issue-dialog";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { Button } from "@/components/ui/button";
import { Drawer, DrawerContent, DrawerTitle } from "@/components/ui/drawer";
import { Sidebar } from "@/components/workspace/sidebar";
import { WorkspaceIcon } from "@/components/workspace/workspace-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useHotkeyScope } from "@/lib/hotkeys/use-hotkey-scope";

export function WorkspaceShell({ children }: { children: ReactNode }) {
  const { canEdit } = useWorkspace();
  const openNewIssue = useNewIssueStore((state) => state.open);

  useHotkeyScope([{ hotkey: "C", callback: () => openNewIssue(), enabled: canEdit }]);

  return (
    <div className="flex h-svh w-full overflow-hidden">
      <aside className="hidden w-60 shrink-0 border-r md:flex">
        <Sidebar />
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <MobileTopBar />
        <main className="flex min-h-0 flex-1 flex-col">{children}</main>
      </div>
      <CommandPanel />
      <NewIssueDialog />
    </div>
  );
}

function MobileTopBar() {
  const { workspace } = useWorkspace();
  const [isOpen, setIsOpen] = useState(false);
  const openCommandPanel = useCommandPanelStore((state) => state.setOpen);
  return (
    <div className="flex h-12 shrink-0 items-center gap-2 border-b px-2 md:hidden">
      <Drawer direction="left" open={isOpen} onOpenChange={setIsOpen}>
        <Button
          variant="ghost"
          size="icon"
          aria-label="Open navigation"
          onClick={() => setIsOpen(true)}
        >
          <MenuIcon className="size-4" />
        </Button>
        <DrawerContent className="w-72 max-w-full">
          <DrawerTitle className="sr-only">Navigation</DrawerTitle>
          <Sidebar onNavigate={() => setIsOpen(false)} />
        </DrawerContent>
      </Drawer>
      <WorkspaceIcon
        id={workspace.id}
        name={workspace.name}
        icon={workspace.icon}
        color={workspace.color}
      />
      <span className="min-w-0 flex-1 truncate text-sm font-medium">{workspace.name}</span>
      <Button
        variant="ghost"
        size="icon"
        aria-label="Search"
        onClick={() => openCommandPanel(true)}
      >
        <SearchIcon className="size-4" />
      </Button>
    </div>
  );
}
