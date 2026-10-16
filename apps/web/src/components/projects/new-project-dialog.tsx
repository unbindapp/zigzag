import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";

import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { PriorityPicker } from "@/components/issues/pickers";
import {
  type TNewProjectDefaults,
  useNewProjectStore,
} from "@/components/projects/new-project-store";
import { type TResolution } from "@/components/projects/project-dates";
import {
  LeadPicker,
  MembersPicker,
  ProjectDatePicker,
  ProjectStatusPicker,
  TeamsPicker,
  useProjectStatuses,
} from "@/components/projects/project-pickers";
import { MarkdownEditor } from "@/components/shared/markdown-editor";
import { type TUserSummary } from "@/components/shared/user-chip";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useTeams } from "@/components/workspace/teams-provider";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  ProjectFormSchema,
  type TProjectStatus,
  useCreateProjectMutation,
} from "@/lib/queries/projects";

export function NewProjectDialog() {
  const isOpen = useNewProjectStore((state) => state.isOpen);
  const close = useNewProjectStore((state) => state.close);
  const defaults = useNewProjectStore((state) => state.defaults);
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && close()}>
      <DialogContent className="w-2xl max-w-full">
        <DialogHeader>
          <DialogTitle>New project</DialogTitle>
        </DialogHeader>
        {isOpen && <NewProjectForm defaults={defaults} onClose={close} />}
      </DialogContent>
    </Dialog>
  );
}

type TTeamRef = { id: string; key: string; name: string; icon: string; color: string };

function useInitialTeams(defaults: TNewProjectDefaults): TTeamRef[] {
  const { teams, myTeams } = useTeams();
  const routeTeam = useTeamFromRoute();
  const id = [defaults.teamId, routeTeam?.id, myTeams[0]?.id, teams[0]?.id].find(Boolean);
  const team = teams.find((candidate) => candidate.id === id);
  return team
    ? [{ id: team.id, key: team.key, name: team.name, icon: team.icon, color: team.color }]
    : [];
}

function NewProjectForm({
  defaults,
  onClose,
}: {
  defaults: TNewProjectDefaults;
  onClose: () => void;
}) {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { teams: allTeams } = useTeams();
  const { statuses } = useProjectStatuses();
  const initialTeams = useInitialTeams(defaults);
  const [teams, setTeams] = useState<TTeamRef[]>(initialTeams);
  const [statusId, setStatusId] = useState<string | undefined>(defaults.statusId);
  const [lead, setLead] = useState<TUserSummary | null>(null);
  const [memberIds, setMemberIds] = useState<string[]>([]);
  const { users } = useWorkspaceUsers();
  const members = users.filter((user) => memberIds.includes(user.id));
  const [priority, setPriority] = useState(0);
  const [target, setTarget] = useState<{ date: string | null; resolution: TResolution }>({
    date: null,
    resolution: "day",
  });
  const [description, setDescription] = useState("");
  const { mutateAsync: createProject } = useCreateProjectMutation();
  const status: TProjectStatus | undefined =
    statuses.find((candidate) => candidate.id === statusId) ??
    statuses.find((candidate) => candidate.type === "backlog");

  const form = useAppForm({
    defaultValues: { name: "", summary: "" },
    validators: { onSubmit: ProjectFormSchema },
    onSubmit: async ({ value }) => {
      const parsed = ProjectFormSchema.parse(value);
      const res = await createProject({
        workspace_id: workspace.id,
        name: parsed.name,
        summary: parsed.summary || undefined,
        description: description || undefined,
        team_ids: teams.map((team) => team.id),
        status_id: status?.id,
        lead_id: lead?.id,
        member_ids: memberIds.length > 0 ? memberIds : undefined,
        priority,
        target_date: target.date ?? undefined,
        target_date_resolution: target.date ? target.resolution : undefined,
        initiative_ids: defaults.initiativeId ? [defaults.initiativeId] : undefined,
      });
      onClose();
      await navigate({
        to: "/$workspace/project/$slug",
        params: { workspace: workspace.slug, slug: res.data.slug },
      });
    },
  });

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField name="name" children={(field) => <field.TextField label="Name" autoFocus />} />
      <form.AppField
        name="summary"
        children={(field) => (
          <field.TextField label="Summary" placeholder="One line about the project" />
        )}
      />
      <MarkdownEditor
        value={description}
        onChange={setDescription}
        placeholder="Describe the project…"
        label="Description"
      />
      <div className="flex flex-wrap items-center gap-2">
        <ProjectStatusPicker
          appearance="compact"
          status={status}
          onChange={(next) => setStatusId(next.id)}
        />
        <TeamsPicker
          appearance="compact"
          teams={teams}
          onChange={(ids) =>
            setTeams(
              allTeams
                .filter((team) => ids.includes(team.id))
                .map((team) => ({
                  id: team.id,
                  key: team.key,
                  name: team.name,
                  icon: team.icon,
                  color: team.color,
                })),
            )
          }
        />
        <LeadPicker appearance="compact" lead={lead} onChange={setLead} />
        <MembersPicker appearance="compact" members={members} onChange={setMemberIds} />
        <PriorityPicker appearance="compact" priority={priority} onChange={setPriority} />
        <ProjectDatePicker
          appearance="compact"
          label="Target date"
          date={target.date}
          resolution={target.resolution}
          onChange={setTarget}
        />
      </div>
      <DialogFooter>
        <DialogClose
          render={
            <Button type="button" variant="outline" size="sm">
              Cancel
            </Button>
          }
        />
        <form.AppForm>
          <form.SubmitButton size="sm" disabled={teams.length === 0}>
            Create project
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
