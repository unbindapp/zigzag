import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  KeyRoundIcon,
  PencilIcon,
  PlusIcon,
  SearchIcon,
  Trash2Icon,
  UsersIcon,
} from "lucide-react";
import { useState } from "react";
import { useDebounceValue } from "usehooks-ts";

import { List, ListRow, ListRowText } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import {
  CreateUserDialog,
  EditUserDialog,
  ResetPasswordDialog,
} from "@/components/settings/user-dialogs";
import { Avatar } from "@/components/shared/avatar";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { meQuery } from "@/lib/queries/me";
import { type TUser, useDeleteUserMutation, usersListQuery } from "@/lib/queries/users";

export const Route = createFileRoute("/$workspace/_workspace/settings/system/users")({
  component: SystemUsersSettingsPage,
});

function SystemUsersSettingsPage() {
  const [search, setSearch] = useState("");
  const [query] = useDebounceValue(search.trim(), 250);
  const { data, isPending, error, refetch } = useQuery(usersListQuery(query));
  const users = data?.users ?? [];
  return (
    <SettingsPage title="Users" description="Every account on this instance.">
      <SettingsSection
        title="Users"
        actions={
          <CreateUserDialog
            trigger={
              <Button size="sm">
                <PlusIcon className="size-4" />
                Create user
              </Button>
            }
          />
        }
      >
        <div className="relative">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute top-2.5 left-3 size-4" />
          <Input
            aria-label="Search users"
            placeholder="Search by name or email"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            className="pl-9"
          />
        </div>
        <ListStates
          isPending={isPending}
          error={error}
          onRetry={refetch}
          isEmpty={users.length === 0}
          empty={{ icon: UsersIcon, title: query ? "No matching users" : "No users yet" }}
        >
          <List>
            {users.map((user) => (
              <UserRow key={user.id} user={user} />
            ))}
          </List>
        </ListStates>
      </SettingsSection>
    </SettingsPage>
  );
}

function UserRow({ user }: { user: TUser }) {
  const { data: me } = useQuery(meQuery);
  const isSelf = me?.user.id === user.id;
  const { mutateAsync: deleteUser } = useDeleteUserMutation();
  const name = user.name || user.display_name || user.email;
  return (
    <ListRow
      actions={
        <>
          <EditUserDialog
            user={user}
            isSelf={isSelf}
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Edit ${name}`}>
                <PencilIcon className="size-4" />
              </Button>
            }
          />
          <ResetPasswordDialog
            user={user}
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Reset password for ${name}`}>
                <KeyRoundIcon className="size-4" />
              </Button>
            }
          />
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Delete ${name}`} disabled={isSelf}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title={`Delete ${name}`}
            description="The account is deleted and removed from every workspace."
            confirmLabel="Delete user"
            confirmText={user.email}
            onConfirm={() => deleteUser({ userId: user.id })}
          />
        </>
      }
    >
      <Avatar id={user.id} name={name} src={user.avatar_url} size="sm" />
      <ListRowText primary={name} secondary={user.email} />
      {user.system_admin && <Badge variant="process">System admin</Badge>}
      {!user.active && <Badge variant="muted">Inactive</Badge>}
    </ListRow>
  );
}
