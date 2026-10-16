import { PlusIcon } from "lucide-react";
import { useState } from "react";
import { z } from "zod";

import { repositoryNameSchema } from "@/components/integrations/github-sync";
import { TeamIcon } from "@/components/shared/team-icon";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useAddGithubRepoSyncMutation } from "@/lib/queries/integrations";

const formSchema = z.object({
  full_name: repositoryNameSchema,
  team_id: z.string().min(1, "Pick a team."),
  sync_closed: z.boolean(),
});

export function AddGithubRepoSyncDialog() {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger
        render={
          <Button size="sm">
            <PlusIcon className="size-4" />
            Add repository
          </Button>
        }
      />
      <DialogContent className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>Add repository</DialogTitle>
          <DialogDescription>
            New GitHub issues land in the team, in a project named after the repository.
          </DialogDescription>
        </DialogHeader>
        {isOpen && <AddGithubRepoSyncForm onDone={() => setIsOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function AddGithubRepoSyncForm({ onDone }: { onDone: () => void }) {
  const { workspace } = useWorkspace();
  const { teams } = useTeams();
  const { mutateAsync: add } = useAddGithubRepoSyncMutation();
  const form = useAppForm({
    defaultValues: { full_name: "", team_id: teams[0]?.id ?? "", sync_closed: false },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      await add({ workspace_id: workspace.id, ...formSchema.parse(value) });
      onDone();
    },
  });
  const teamItems = teams.map((team) => ({
    value: team.id,
    label: team.name,
    icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
  }));
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="full_name"
        children={(field) => (
          <field.TextField label="Repository" placeholder="unbindapp/zigzag" autoFocus />
        )}
      />
      <form.AppField
        name="team_id"
        children={(field) => <field.SelectField label="Team" items={teamItems} />}
      />
      <form.AppField
        name="sync_closed"
        children={(field) => (
          <field.SwitchField
            label="Include closed issues"
            description="Imports of existing issues bring closed ones too."
          />
        )}
      />
      <DialogFooter>
        <DialogClose
          render={
            <Button type="button" variant="outline" size="sm">
              Cancel
            </Button>
          }
        />
        <form.AppForm>
          <form.SubmitButton size="sm">Add repository</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
