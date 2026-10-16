import { useQuery } from "@tanstack/react-query";
import { useNavigate, useRouter } from "@tanstack/react-router";
import {
  CheckIcon,
  ChevronDownIcon,
  LogOutIcon,
  PlusIcon,
  ServerIcon,
  SettingsIcon,
  UserIcon,
} from "lucide-react";

import { WorkspaceIcon } from "@/components/workspace/workspace-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useLogoutMutation } from "@/lib/queries/auth";
import { meQuery, meWorkspaces } from "@/lib/queries/me";

export function WorkspaceSwitcher({ onNavigate }: { onNavigate?: () => void }) {
  const { workspace, isAdmin, isSystemAdmin } = useWorkspace();
  const { data: me } = useQuery(meQuery);
  const navigate = useNavigate();
  const router = useRouter();
  const { mutate: signOut } = useLogoutMutation();
  const workspaces = meWorkspaces(me);

  const go = (to: () => void) => {
    to();
    onNavigate?.();
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="ghost"
            size="sm"
            forceMinSize={false}
            aria-label="Switch workspace"
            className="min-w-0 flex-1 justify-start"
          >
            <WorkspaceIcon
              id={workspace.id}
              name={workspace.name}
              icon={workspace.icon}
              color={workspace.color}
              iconUrl={workspace.icon_url}
            />
            <span className="min-w-0 truncate">{workspace.name}</span>
            <ChevronDownIcon className="text-muted-foreground size-3.5" />
          </Button>
        }
      />
      <DropdownMenuContent className="w-64">
        <DropdownMenuGroup>
          <DropdownMenuLabel>{me?.user.email}</DropdownMenuLabel>
          {workspaces.map((item) => (
            <DropdownMenuItem
              key={item.id}
              onClick={() =>
                go(() => navigate({ to: "/$workspace", params: { workspace: item.slug } }))
              }
            >
              <WorkspaceIcon
                id={item.id}
                name={item.name}
                icon={item.icon}
                color={item.color}
                iconUrl={item.icon_url}
              />
              <span className="min-w-0 flex-1 truncate">{item.name}</span>
              {item.id === workspace.id && <CheckIcon className="size-4" />}
            </DropdownMenuItem>
          ))}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          {isAdmin && (
            <DropdownMenuItem
              onClick={() =>
                go(() =>
                  navigate({ to: "/$workspace/settings", params: { workspace: workspace.slug } }),
                )
              }
            >
              <SettingsIcon className="size-4" />
              Workspace settings
            </DropdownMenuItem>
          )}
          {isSystemAdmin && (
            <DropdownMenuItem
              onClick={() =>
                go(() =>
                  navigate({
                    to: "/$workspace/settings/system",
                    params: { workspace: workspace.slug },
                  }),
                )
              }
            >
              <ServerIcon className="size-4" />
              System settings
            </DropdownMenuItem>
          )}
          <DropdownMenuItem
            onClick={() =>
              go(() =>
                navigate({
                  to: "/$workspace/settings/account",
                  params: { workspace: workspace.slug },
                }),
              )
            }
          >
            <UserIcon className="size-4" />
            Account settings
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => go(() => navigate({ to: "/create-workspace" }))}>
            <PlusIcon className="size-4" />
            Create workspace
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem
            onClick={() => signOut(undefined, { onSuccess: () => router.history.push("/sign-in") })}
          >
            <LogOutIcon className="size-4" />
            Sign out
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
