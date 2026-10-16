import { useQuery } from "@tanstack/react-query";
import { type LinkProps, useNavigate } from "@tanstack/react-router";
import {
  BoxIcon,
  CircleDotIcon,
  FileTextIcon,
  InboxIcon,
  LayersIcon,
  ListTodoIcon,
  SettingsIcon,
  SquarePenIcon,
  SunMoonIcon,
  TargetIcon,
  UserCircleIcon,
  UsersRoundIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import { ReactNode, useState } from "react";

import { useCommandPanelStore } from "@/components/command-panel/command-panel-store";
import { useCommandPanelSearch } from "@/components/command-panel/use-command-panel-search";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { useNewProjectStore } from "@/components/projects/new-project-store";
import { nextTheme } from "@/components/providers/themes";
import { useCreateTeamStore } from "@/components/settings/create-team-store";
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
import { useHotkeyScope } from "@/lib/hotkeys/use-hotkey-scope";
import { meQuery, meWorkspaces } from "@/lib/queries/me";

export function CommandPanel() {
  const isOpen = useCommandPanelStore((state) => state.isOpen);
  const setOpen = useCommandPanelStore((state) => state.setOpen);

  useHotkeyScope([{ hotkey: "Mod+K", callback: () => setOpen(true), ignoreInputs: false }]);
  useHotkeyScope([{ hotkey: "Mod+K", callback: () => setOpen(false), ignoreInputs: false }], {
    enabled: isOpen,
    modal: true,
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
  const { theme: currentTheme, setTheme } = useTheme();
  const { workspace, isAdmin, canEdit } = useWorkspace();
  const { myTeams } = useTeams();
  const { data: me } = useQuery(meQuery);
  const openNewIssue = useNewIssueStore((state) => state.open);
  const openNewProject = useNewProjectStore((state) => state.open);
  const openCreateTeam = useCreateTeamStore((state) => state.setOpen);
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
          {
            id: "new-project",
            label: "Create project",
            icon: <BoxIcon className="size-4" />,
            run: run(() => openNewProject()),
          },
        ]
      : []),
    ...(isAdmin
      ? [
          {
            id: "new-team",
            label: "Create team",
            icon: <UsersRoundIcon className="size-4" />,
            run: run(() => openCreateTeam(true)),
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
      icon: (
        <WorkspaceIcon
          id={item.id}
          name={item.name}
          icon={item.icon}
          color={item.color}
          iconUrl={item.icon_url}
        />
      ),
      run: () => go({ to: "/$workspace", params: { workspace: item.slug } }),
    }));

  const theme: TAction[] = [
    {
      id: "theme-toggle",
      label: "Toggle theme",
      keywords: ["theme", "appearance", "light", "dark", "system"],
      icon: <SunMoonIcon className="size-4" />,
      run: run(() => setTheme(nextTheme(currentTheme))),
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
