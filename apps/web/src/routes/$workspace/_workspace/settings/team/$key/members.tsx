import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Trash2Icon, UserPlusIcon, UsersIcon } from "lucide-react";
import { useState } from "react";

import { List, ListRow, ListRowText } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { Avatar } from "@/components/shared/avatar";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import {
  type TTeam,
  type TTeamMember,
  teamMembersQuery,
  useAddTeamMemberMutation,
  useRemoveTeamMemberMutation,
} from "@/lib/queries/teams";
import { workspaceMembersQuery } from "@/lib/queries/workspaces";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/members")({
  component: TeamMembersSettingsPage,
});

function TeamMembersSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { data, isPending, error, refetch } = useQuery(teamMembersQuery(team.id));
  const members = data?.members ?? [];
  return (
    <SettingsPage title="Members" description="People who belong to this team.">
      <SettingsSection title="Add member">
        <AddMember team={team} memberIds={members.map((member) => member.user.id)} />
      </SettingsSection>
      <SettingsSection title="Members">
        <ListStates
          isPending={isPending}
          error={error}
          onRetry={refetch}
          isEmpty={members.length === 0}
          empty={{ icon: UsersIcon, title: "No members yet" }}
        >
          <List>
            {members.map((member) => (
              <TeamMemberRow key={member.user.id} team={team} member={member} />
            ))}
          </List>
        </ListStates>
      </SettingsSection>
    </SettingsPage>
  );
}

function AddMember({ team, memberIds }: { team: TTeam; memberIds: string[] }) {
  const { workspace } = useWorkspace();
  const { data } = useQuery(workspaceMembersQuery(workspace.id));
  const [userId, setUserId] = useState<string | null>(null);
  const { mutate: addMember, isPending } = useAddTeamMemberMutation();
  const candidates = (data?.members ?? []).filter((member) => !memberIds.includes(member.user.id));
  const items = candidates.map((member) => ({
    value: member.user.id,
    label: member.user.name || member.user.display_name,
  }));

  if (data && candidates.length === 0) {
    return (
      <p className="text-muted-foreground text-sm">Everyone in the workspace is in this team.</p>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <Select items={items} value={userId} onValueChange={(value) => setUserId(value)}>
        <SelectTrigger aria-label="Workspace member" className="min-w-0 flex-1">
          <SelectValue placeholder="Pick a workspace member" />
        </SelectTrigger>
        <SelectContent>
          {candidates.map((member) => (
            <SelectItem key={member.user.id} value={member.user.id}>
              <Avatar
                id={member.user.id}
                name={member.user.name}
                src={member.user.avatar_url}
                size="xs"
              />
              <span className="truncate">{member.user.name || member.user.display_name}</span>
              <span className="text-muted-foreground truncate text-xs">{member.user.email}</span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button
        disabled={!userId}
        isPending={isPending}
        onClick={() =>
          userId && addMember({ teamId: team.id, userId }, { onSuccess: () => setUserId(null) })
        }
      >
        <UserPlusIcon className="size-4" />
        Add
      </Button>
    </div>
  );
}

function TeamMemberRow({ team, member }: { team: TTeam; member: TTeamMember }) {
  const { mutateAsync: removeMember } = useRemoveTeamMemberMutation();
  const name = member.user.name || member.user.display_name;
  return (
    <ListRow
      actions={
        <ConfirmDialog
          trigger={
            <Button variant="ghost" size="icon" aria-label={`Remove ${name} from ${team.name}`}>
              <Trash2Icon className="size-4" />
            </Button>
          }
          title={`Remove ${name}`}
          description={`${name} leaves ${team.name} but stays in the workspace.`}
          confirmLabel="Remove"
          onConfirm={() => removeMember({ teamId: team.id, userId: member.user.id })}
        />
      }
    >
      <Avatar id={member.user.id} name={name} src={member.user.avatar_url} size="sm" />
      <ListRowText
        primary={name}
        secondary={member.user.display_name ? `@${member.user.display_name}` : undefined}
      />
      {member.owner && <Badge variant="muted">Owner</Badge>}
    </ListRow>
  );
}
