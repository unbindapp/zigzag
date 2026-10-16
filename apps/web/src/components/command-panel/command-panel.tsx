import { useHotkey } from "@tanstack/react-hotkeys";
import { useQuery } from "@tanstack/react-query";
import { type LinkProps, useNavigate } from "@tanstack/react-router";
import {
  BoxIcon,
  CircleDotIcon,
  FileTextIcon,
  InboxIcon,
  LayersIcon,
  ListTodoIcon,
  MonitorIcon,
  MoonIcon,
  SettingsIcon,
  SquarePenIcon,
  SunIcon,
  TargetIcon,
  UserCircleIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import { ReactNode, useState } from "react";

import { useCommandPanelStore } from "@/components/command-panel/command-panel-store";
import { useCommandPanelSearch } from "@/components/command-panel/use-command-panel-search";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { TeamIcon } from "@/components/shared/team-icon";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from "@/components/ui/command";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useTeams } from "@/components/workspace/teams-provider";
import { WorkspaceIcon } from "@/components/workspace/workspace-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { meQuery, meWorkspaces } from "@/lib/queries/me";

export function CommandPanel() {
  const isOpen = useCommandPanelStore((state) => state.isOpen);
  const setOpen = useCommandPanelStore((state) => state.setOpen);

  useHotkey("Mod+K", () => setOpen(!useCommandPanelStore.getState().isOpen), {
    conflictBehavior: "allow",
    ignoreInputs: false,
  });

  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      <DialogContent
        variant="styleless"
        hideXButton
        animate={false}
        className="w-2xl max-w-full self-start"
      >
        <DialogTitle className="sr-only">Command panel</DialogTitle>
        {isOpen && <CommandPanelContent onDone={() => setOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

type TAction = {
  id: string;
  label: string;
  icon: ReactNode;
  keywords?: string[];
  shortcut?: string;
  run: () => void;
};

function CommandPanelContent({ onDone }: { onDone: () => void }) {
  const [query, setQuery] = useState("");
  const navigate = useNavigate();
  const { setTheme } = useTheme();
  const { workspace, isAdmin, canEdit } = useWorkspace();
  const { myTeams } = useTeams();
  const { data: me } = useQuery(meQuery);
  const openNewIssue = useNewIssueStore((state) => state.open);
  const search = useCommandPanelSearch(query);
  const params = { workspace: workspace.slug };

  const go = (link: LinkProps) => {
    onDone();
    void navigate(link);
  };
  const run = (fn: () => void) => () => {
    onDone();
    fn();
  };

  const actions: TAction[] = [
    ...(canEdit
      ? [
          {
            id: "new-issue",
            label: "New issue",
            icon: <SquarePenIcon className="size-4" />,
            shortcut: "C",
            run: run(() => openNewIssue()),
          },
        ]
      : []),
    {
      id: "go-inbox",
      label: "Go to inbox",
      icon: <InboxIcon className="size-4" />,
      run: () => go({ to: "/$workspace/inbox", params }),
    },
    {
      id: "go-my-issues",
      label: "Go to my issues",
      icon: <UserCircleIcon className="size-4" />,
      run: () => go({ to: "/$workspace/my-issues", params }),
    },
    {
      id: "go-settings",
      label: "Go to settings",
      icon: <SettingsIcon className="size-4" />,
      run: () =>
        go({ to: isAdmin ? "/$workspace/settings" : "/$workspace/settings/account", params }),
    },
  ];

  const navigation: TAction[] = [
    {
      id: "nav-issues",
      label: "All issues",
      icon: <ListTodoIcon className="size-4" />,
      run: () => go({ to: "/$workspace/issues", params }),
    },
    {
      id: "nav-projects",
      label: "Projects",
      icon: <BoxIcon className="size-4" />,
      run: () => go({ to: "/$workspace/projects", params }),
    },
    {
      id: "nav-views",
      label: "Views",
      icon: <LayersIcon className="size-4" />,
      run: () => go({ to: "/$workspace/views", params }),
    },
    {
      id: "nav-initiatives",
      label: "Initiatives",
      icon: <TargetIcon className="size-4" />,
      run: () => go({ to: "/$workspace/initiatives", params }),
    },
    {
      id: "nav-documents",
      label: "Documents",
      icon: <FileTextIcon className="size-4" />,
      run: () => go({ to: "/$workspace/documents", params }),
    },
    ...myTeams.map((team) => ({
      id: `nav-team-${team.id}`,
      label: `${team.name} issues`,
      keywords: [team.key, "team"],
      icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
      run: () => go({ to: "/$workspace/team/$key/issues", params: { ...params, key: team.key } }),
    })),
  ];

  const workspaces: TAction[] = meWorkspaces(me)
    .filter((item) => item.id !== workspace.id)
    .map((item) => ({
      id: `ws-${item.id}`,
      label: `Switch to ${item.name}`,
      keywords: [item.slug, "workspace"],
      icon: <WorkspaceIcon id={item.id} name={item.name} icon={item.icon} color={item.color} />,
      run: () => go({ to: "/$workspace", params: { workspace: item.slug } }),
    }));

  const theme: TAction[] = [
    {
      id: "theme-light",
      label: "Switch to light theme",
      keywords: ["theme", "appearance"],
      icon: <SunIcon className="size-4" />,
      run: run(() => setTheme("light")),
    },
    {
      id: "theme-dark",
      label: "Switch to dark theme",
      keywords: ["theme", "appearance"],
      icon: <MoonIcon className="size-4" />,
      run: run(() => setTheme("dark")),
    },
    {
      id: "theme-system",
      label: "Use system theme",
      keywords: ["theme", "appearance"],
      icon: <MonitorIcon className="size-4" />,
      run: run(() => setTheme("system")),
    },
  ];

  return (
    <Command loop>
      <CommandInput
        autoFocus
        value={query}
        onValueChange={setQuery}
        showSpinner={search.isPending}
        placeholder="Search or type a command…"
      />
      <ScrollArea className="min-h-0">
        <CommandList>
          <CommandEmpty>No results</CommandEmpty>
          {search.groups.map((group) => (
            <CommandGroup key={group.heading} heading={group.heading}>
              {group.items.map((item) => (
                <CommandItem
                  key={item.id}
                  value={`search-${item.id}`}
                  keywords={[item.label, item.description ?? "", query]}
                  onSelect={() => go(item.link)}
                >
                  {item.icon ?? <CircleDotIcon className="size-4" />}
                  <span className="truncate">{item.label}</span>
                  {item.description && (
                    <span className="text-muted-foreground ml-auto truncate text-xs">
                      {item.description}
                    </span>
                  )}
                </CommandItem>
              ))}
            </CommandGroup>
          ))}
          <ActionGroup heading="Actions" actions={actions} />
          <ActionGroup heading="Navigation" actions={navigation} />
          {workspaces.length > 0 && <ActionGroup heading="Workspaces" actions={workspaces} />}
          <ActionGroup heading="Theme" actions={theme} />
        </CommandList>
      </ScrollArea>
    </Command>
  );
}

function ActionGroup({ heading, actions }: { heading: string; actions: TAction[] }) {
  return (
    <CommandGroup heading={heading}>
      {actions.map((action) => (
        <CommandItem
          key={action.id}
          value={action.id}
          keywords={[action.label, ...(action.keywords ?? [])]}
          onSelect={action.run}
        >
          {action.icon}
          <span className="truncate">{action.label}</span>
          {action.shortcut && <CommandShortcut>{action.shortcut}</CommandShortcut>}
        </CommandItem>
      ))}
    </CommandGroup>
  );
}
