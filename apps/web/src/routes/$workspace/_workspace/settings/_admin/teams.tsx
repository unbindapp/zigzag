import { createFileRoute } from "@tanstack/react-router";
import { LockIcon, PlusIcon, SettingsIcon, Trash2Icon, UsersRoundIcon } from "lucide-react";

import { List, ListRow, ListRowText } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { TeamIcon } from "@/components/shared/team-icon";
import { CreateTeamDialog } from "@/components/settings/create-team-dialog";
import { Button, LinkButton } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TTeam, useDeleteTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/teams")({
  component: WorkspaceTeamsSettingsPage,
});

function WorkspaceTeamsSettingsPage() {
  const { teams, isPending, error, refetch } = useTeams();
  const createButton = (
    <CreateTeamDialog
      trigger={
        <Button size="sm">
          <PlusIcon className="size-4" />
          Create team
        </Button>
      }
    />
  );

  return (
    <SettingsPage title="Teams" description="Teams own issues, workflows and cycles.">
      <SettingsSection title="Teams" actions={createButton}>
        <ListStates
          isPending={isPending}
          error={error}
          onRetry={refetch}
          isEmpty={teams.length === 0}
          empty={{
            icon: UsersRoundIcon,
            title: "No teams yet",
            description: "Create a team to start tracking issues.",
          }}
        >
          <List>
            {teams.map((team) => (
              <TeamRow key={team.id} team={team} />
            ))}
          </List>
        </ListStates>
      </SettingsSection>
    </SettingsPage>
  );
}

function TeamRow({ team }: { team: TTeam }) {
  const { workspace } = useWorkspace();
  const { mutateAsync: deleteTeam } = useDeleteTeamMutation();
  return (
    <ListRow
      actions={
        <>
          <LinkButton
            to="/$workspace/settings/team/$key"
            params={{ workspace: workspace.slug, key: team.key }}
            variant="ghost"
            size="icon"
            aria-label={`${team.name} settings`}
          >
            <SettingsIcon className="size-4" />
          </LinkButton>
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Delete ${team.name}`}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title={`Delete ${team.name}`}
            description="Deletes the team with all of its issues, states and cycles."
            confirmLabel="Delete team"
            confirmText={team.key}
            onConfirm={() => deleteTeam({ teamId: team.id })}
          />
        </>
      }
    >
      <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} />
      <ListRowText
        primary={team.name}
        secondary={`${team.member_count} ${team.member_count === 1 ? "member" : "members"}`}
      />
      <span className="text-muted-foreground font-mono text-xs">{team.key}</span>
      {team.private && (
        <SimpleTooltip content="Private team">
          <LockIcon aria-label="Private team" className="text-muted-foreground size-4" />
        </SimpleTooltip>
      )}
    </ListRow>
  );
}
