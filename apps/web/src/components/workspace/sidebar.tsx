import { useQuery } from "@tanstack/react-query";
import {
  BoxIcon,
  CircleDotIcon,
  CircleHelpIcon,
  FileTextIcon,
  InboxIcon,
  LayersIcon,
  ListTodoIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  SettingsIcon,
  SquarePenIcon,
  TargetIcon,
  TriangleIcon,
  UserCircleIcon,
} from "lucide-react";

import { useCommandPanelStore } from "@/components/command-panel/command-panel-store";
import { SidebarFavorites } from "@/components/favorites/sidebar-favorites";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { useCreateTeamStore } from "@/components/settings/create-team-store";
import { TeamIcon } from "@/components/shared/team-icon";
import { Badge } from "@/components/ui/badge";
import { Button, LinkButton } from "@/components/ui/button";
import { Kbd } from "@/components/ui/kbd";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import {
  SidebarCollapsible,
  SidebarNavItem,
  SidebarSection,
} from "@/components/workspace/sidebar-nav";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { WorkspaceSwitcher } from "@/components/workspace/workspace-switcher";
import { notificationCountsQuery } from "@/lib/queries/notifications";
import { type TTeam } from "@/lib/queries/teams";

type TProps = {
  onNavigate?: () => void;
};

export function Sidebar({ onNavigate }: TProps) {
  const { workspace, canEdit, isAdmin } = useWorkspace();
  const openCommandPanel = useCommandPanelStore((state) => state.setOpen);
  const openNewIssue = useNewIssueStore((state) => state.open);
  const openCreateTeam = useOpenCreateTeam(onNavigate);
  const params = { workspace: workspace.slug };

  return (
    <div className="bg-card flex h-full w-full flex-col">
      <div className="flex items-center gap-2 p-2">
        <WorkspaceSwitcher onNavigate={onNavigate} />
        {canEdit && (
          <Button
            variant="outline"
            size="icon"
            aria-label="New issue"
            onClick={() => {
              onNavigate?.();
              openNewIssue();
            }}
          >
            <SquarePenIcon className="size-4" />
          </Button>
        )}
      </div>
      <div className="px-2 pb-2">
        <Button
          variant="outline"
          size="sm"
          forceMinSize={false}
          onClick={() => {
            onNavigate?.();
            openCommandPanel(true);
          }}
          className="text-muted-foreground w-full justify-start font-normal"
        >
          <SearchIcon className="size-4" />
          <span className="flex-1 text-left">Search</span>
          <Kbd>⌘K</Kbd>
        </Button>
      </div>
      <ScrollArea className="min-h-0" classNameViewport="px-2 pb-2">
        <nav aria-label="Workspace" className="flex flex-col gap-4">
          <div className="flex flex-col gap-0.5">
            <SidebarNavItem
              label="Inbox"
              icon={InboxIcon}
              link={{ to: "/$workspace/inbox", params }}
              badge={<InboxUnreadBadge />}
              onNavigate={onNavigate}
            />
            <SidebarNavItem
              label="My issues"
              icon={UserCircleIcon}
              link={{ to: "/$workspace/my-issues", params }}
              onNavigate={onNavigate}
            />
          </div>
          <SidebarSection title="Workspace">
            <SidebarNavItem
              label="Projects"
              icon={BoxIcon}
              link={{ to: "/$workspace/projects", params }}
              onNavigate={onNavigate}
            />
            <SidebarNavItem
              label="Views"
              icon={LayersIcon}
              link={{ to: "/$workspace/views", params }}
              onNavigate={onNavigate}
            />
            <SidebarNavItem
              label="Initiatives"
              icon={TargetIcon}
              link={{ to: "/$workspace/initiatives", params }}
              onNavigate={onNavigate}
            />
            <SidebarNavItem
              label="Documents"
              icon={FileTextIcon}
              link={{ to: "/$workspace/documents", params }}
              onNavigate={onNavigate}
            />
            <SidebarNavItem
              label="All issues"
              icon={ListTodoIcon}
              link={{ to: "/$workspace/issues", params }}
              onNavigate={onNavigate}
            />
          </SidebarSection>
          <SidebarSection title="Favorites">
            <SidebarFavorites onNavigate={onNavigate} />
          </SidebarSection>
          <SidebarSection
            title="Your teams"
            action={
              isAdmin && (
                <Button
                  variant="ghost"
                  size="icon"
                  forceMinSize={false}
                  aria-label="New team"
                  onClick={openCreateTeam}
                  className="text-muted-foreground"
                >
                  <PlusIcon className="size-3.5" />
                </Button>
              )
            }
          >
            <TeamsNav onNavigate={onNavigate} />
          </SidebarSection>
        </nav>
      </ScrollArea>
      <SidebarFooter onNavigate={onNavigate} />
    </div>
  );
}

function InboxUnreadBadge() {
  const { workspace } = useWorkspace();
  const { data } = useQuery(notificationCountsQuery(workspace.id));
  const unread = data?.unread ?? 0;
  if (unread === 0) return null;
  return (
    <Badge variant="muted" aria-label={`${unread} unread`}>
      {unread > 99 ? "99+" : unread}
    </Badge>
  );
}

function useOpenCreateTeam(onNavigate?: () => void) {
  const setOpen = useCreateTeamStore((state) => state.setOpen);
  return () => {
    onNavigate?.();
    setOpen(true);
  };
}

function TeamsNav({ onNavigate }: TProps) {
  const { isAdmin } = useWorkspace();
  const { myTeams, isPending, error, refetch } = useTeams();
  const openCreateTeam = useOpenCreateTeam(onNavigate);
  if (isPending) {
    return (
      <div className="flex flex-col gap-2 px-2.5 py-1">
        <Skeleton className="h-3 w-2/3" />
        <Skeleton className="h-3 w-1/2" />
      </div>
    );
  }
  if (error) {
    return (
      <Button
        variant="ghost"
        size="sm"
        onClick={refetch}
        className="text-destructive justify-start"
      >
        Could not load teams. Retry
      </Button>
    );
  }
  if (myTeams.length === 0 && isAdmin) {
    return (
      <Button
        variant="ghost"
        size="sm"
        onClick={openCreateTeam}
        className="text-muted-foreground justify-start font-normal"
      >
        <PlusIcon className="size-4" />
        Create a team
      </Button>
    );
  }
  if (myTeams.length === 0) {
    return <p className="text-muted-foreground px-2.5 text-xs">You are not in any team yet.</p>;
  }
  return myTeams.map((team) => <TeamNav key={team.id} team={team} onNavigate={onNavigate} />);
}

function TeamNav({ team, onNavigate }: { team: TTeam } & TProps) {
  const { workspace } = useWorkspace();
  const params = { workspace: workspace.slug, key: team.key };
  return (
    <SidebarCollapsible
      label={team.name}
      icon={<TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />}
    >
      <SidebarNavItem
        label="Issues"
        icon={CircleDotIcon}
        link={{ to: "/$workspace/team/$key/issues", params }}
        onNavigate={onNavigate}
      />
      <SidebarNavItem
        label="Backlog"
        icon={ListTodoIcon}
        link={{ to: "/$workspace/team/$key/backlog", params }}
        onNavigate={onNavigate}
      />
      {team.triage_enabled && (
        <SidebarNavItem
          label="Triage"
          icon={TriangleIcon}
          link={{ to: "/$workspace/team/$key/triage", params }}
          onNavigate={onNavigate}
        />
      )}
      {team.cycles_enabled && (
        <SidebarNavItem
          label="Cycles"
          icon={RefreshCwIcon}
          link={{ to: "/$workspace/team/$key/cycles", params }}
          onNavigate={onNavigate}
        />
      )}
      <SidebarNavItem
        label="Projects"
        icon={BoxIcon}
        link={{ to: "/$workspace/team/$key/projects", params }}
        onNavigate={onNavigate}
      />
    </SidebarCollapsible>
  );
}

function SidebarFooter({ onNavigate }: TProps) {
  const { workspace, isAdmin } = useWorkspace();
  const params = { workspace: workspace.slug };
  return (
    <div className="flex items-center gap-2 border-t p-2">
      <LinkButton
        to={isAdmin ? "/$workspace/settings" : "/$workspace/settings/account"}
        params={params}
        variant="ghost"
        size="icon"
        aria-label="Settings"
        onClick={onNavigate}
      >
        <SettingsIcon className="size-4" />
      </LinkButton>
      <Popover>
        <PopoverTrigger
          render={
            <Button variant="ghost" size="icon" aria-label="Help">
              <CircleHelpIcon className="size-4" />
            </Button>
          }
        />
        <PopoverContent side="top" align="start" className="flex w-64 flex-col gap-2 p-3">
          <p className="text-sm font-medium">Keyboard shortcuts</p>
          <ShortcutRow label="Open command panel" keys="⌘K" />
          <ShortcutRow label="New issue" keys="C" />
        </PopoverContent>
      </Popover>
    </div>
  );
}

function ShortcutRow({ label, keys }: { label: string; keys: string }) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <span className="text-muted-foreground flex-1">{label}</span>
      <Kbd>{keys}</Kbd>
    </div>
  );
}
