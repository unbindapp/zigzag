import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import type { ReactNode } from "react";
import {
  LinkIcon,
  MailIcon,
  RotateCwIcon,
  Trash2Icon,
  UserPlusIcon,
  UsersIcon,
} from "lucide-react";

import { List, ListRow, ListRowText } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { useInvitationLinksStore } from "@/components/settings/invitation-links-store";
import { InviteMemberDialog } from "@/components/settings/invite-member-dialog";
import { Avatar } from "@/components/shared/avatar";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { RelativeTime } from "@/components/shared/relative-time";
import { RoleSelect } from "@/components/shared/role-select";
import { TeamIcon } from "@/components/shared/team-icon";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { meQuery } from "@/lib/queries/me";
import {
  invitationsListQuery,
  type TInvitation,
  type TWorkspaceMember,
  useRemoveWorkspaceMemberMutation,
  useResendInvitationMutation,
  useRevokeInvitationMutation,
  useUpdateWorkspaceMemberMutation,
  workspaceMembersQuery,
  workspaceRoleLabels,
} from "@/lib/queries/workspaces";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/members")({
  component: WorkspaceMembersSettingsPage,
});

function WorkspaceMembersSettingsPage() {
  const inviteButton = (
    <InviteMemberDialog
      trigger={
        <Button size="sm">
          <UserPlusIcon className="size-4" />
          Invite
        </Button>
      }
    />
  );
  return (
    <SettingsPage title="Members" description="People in this workspace and their roles.">
      <MembersSection actions={inviteButton} />
      <InvitationsSection />
    </SettingsPage>
  );
}

function MembersSection({ actions }: { actions: ReactNode }) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(workspaceMembersQuery(workspace.id));
  const members = data?.members ?? [];
  return (
    <SettingsSection title="Members" actions={actions}>
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={refetch}
        isEmpty={members.length === 0}
        empty={{ icon: UsersIcon, title: "No members yet" }}
      >
        <List>
          {members.map((member) => (
            <MemberRow key={member.user.id} member={member} />
          ))}
        </List>
      </ListStates>
    </SettingsSection>
  );
}

function MemberRow({ member }: { member: TWorkspaceMember }) {
  const { workspace } = useWorkspace();
  const { teams } = useTeams();
  const { data: me } = useQuery(meQuery);
  const isSelf = me?.user.id === member.user.id;
  const { mutate: updateRole, isPending: isUpdating } = useUpdateWorkspaceMemberMutation(
    workspace.id,
  );
  const { mutateAsync: remove } = useRemoveWorkspaceMemberMutation(workspace.id);
  const memberTeams = teams.filter((team) => member.team_ids?.includes(team.id));
  const name = member.user.name || member.user.display_name;

  return (
    <ListRow
      actions={
        <>
          <div className="hidden items-center gap-1.5 sm:flex">
            {memberTeams.slice(0, 3).map((team) => (
              <SimpleTooltip key={team.id} content={team.name}>
                <span>
                  <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />
                </span>
              </SimpleTooltip>
            ))}
            {memberTeams.length > 3 && (
              <span className="text-muted-foreground text-xs">+{memberTeams.length - 3}</span>
            )}
          </div>
          <RoleSelect
            value={member.role}
            label={`Role for ${name}`}
            disabled={isSelf || isUpdating}
            onChange={(role) =>
              updateRole({ workspaceId: workspace.id, userId: member.user.id, role })
            }
          />
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Remove ${name}`} disabled={isSelf}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title={`Remove ${name}`}
            description="They lose access to this workspace and leave all of its teams."
            confirmLabel="Remove member"
            onConfirm={() => remove({ workspaceId: workspace.id, userId: member.user.id })}
          />
        </>
      }
    >
      <Avatar id={member.user.id} name={name} src={member.user.avatar_url} size="sm" />
      <ListRowText primary={name} secondary={member.user.email} />
      {!member.user.active && <Badge variant="muted">Inactive</Badge>}
    </ListRow>
  );
}

function InvitationsSection() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(invitationsListQuery(workspace.id));
  const invitations = data?.invitations ?? [];
  return (
    <SettingsSection
      title="Pending invitations"
      description="Invitations that have not been accepted yet."
    >
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={refetch}
        isEmpty={invitations.length === 0}
        skeletonRows={2}
        empty={{ icon: MailIcon, title: "No pending invitations yet" }}
      >
        <List>
          {invitations.map((invitation) => (
            <InvitationRow key={invitation.id} invitation={invitation} />
          ))}
        </List>
      </ListStates>
    </SettingsSection>
  );
}

function InvitationRow({ invitation }: { invitation: TInvitation }) {
  const { workspace } = useWorkspace();
  const links = useInvitationLinksStore((state) => state.links);
  const remember = useInvitationLinksStore((state) => state.remember);
  const { mutateAsync: resend, isPending: isResending } = useResendInvitationMutation(workspace.id);
  const { mutateAsync: revoke } = useRevokeInvitationMutation(workspace.id);
  const input = { workspaceId: workspace.id, invitationId: invitation.id };

  const resendInvitation = async () => {
    const res = await resend(input).catch(() => null);
    if (!res) return null;
    if (res.data.link) remember(invitation.id, res.data.link);
    return res;
  };

  const copyLink = async () => {
    const known = links[invitation.id];
    const link = known ?? (await resendInvitation())?.data.link;
    if (!link) return;
    await navigator.clipboard.writeText(link);
    toast.add({ type: "success", title: known ? "Link copied" : "New invitation link copied" });
  };

  return (
    <ListRow
      actions={
        <>
          <Button
            variant="ghost"
            size="icon"
            aria-label={`Copy invitation link for ${invitation.email}`}
            onClick={copyLink}
          >
            <LinkIcon className="size-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            aria-label={`Resend invitation to ${invitation.email}`}
            isPending={isResending}
            onClick={async () => {
              const res = await resendInvitation();
              if (!res) return;
              toast.add({
                type: "success",
                title: res.data.emailed ? "Invitation resent" : "New link created",
                description: res.data.emailed
                  ? undefined
                  : "Email is not set up. Copy the link to share it.",
              });
            }}
          >
            <RotateCwIcon className="size-4" />
          </Button>
          <ConfirmDialog
            trigger={
              <Button
                variant="ghost"
                size="icon"
                aria-label={`Revoke invitation for ${invitation.email}`}
              >
                <Trash2Icon className="size-4" />
              </Button>
            }
            title="Revoke invitation"
            description={`${invitation.email} can no longer use the invitation link.`}
            confirmLabel="Revoke invitation"
            onConfirm={() => revoke(input)}
          />
        </>
      }
    >
      <MailIcon className="text-muted-foreground size-4" />
      <ListRowText primary={invitation.email} secondary={workspaceRoleLabels[invitation.role]} />
      <span className="text-muted-foreground hidden text-xs sm:inline">
        Expires <RelativeTime value={invitation.expires_at} />
      </span>
    </ListRow>
  );
}
