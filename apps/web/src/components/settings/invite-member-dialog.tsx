import { ReactElement, useState } from "react";
import { z } from "zod";

import { useInvitationLinksStore } from "@/components/settings/invitation-links-store";
import { CheckboxList } from "@/components/shared/checkbox-list";
import { CopyField } from "@/components/shared/copy-button";
import { roleSelectItems } from "@/components/shared/role-select";
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
import { toast } from "@/components/ui/toast";
import { useTeams } from "@/components/workspace/teams-provider";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  type TCreatedInvitation,
  type TWorkspaceRole,
  useCreateInvitationMutation,
} from "@/lib/queries/workspaces";

const formSchema = z.object({
  email: z.string().trim().email("Enter a valid email address."),
  role: z.enum(["admin", "member", "viewer"]),
  team_ids: z.array(z.string()),
});

export function InviteMemberDialog({ trigger }: { trigger: ReactElement }) {
  const [isOpen, setIsOpen] = useState(false);
  const [created, setCreated] = useState<TCreatedInvitation | null>(null);
  const setOpen = (open: boolean) => {
    setIsOpen(open);
    if (!open) setCreated(null);
  };
  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      <DialogTrigger render={trigger} />
      <DialogContent className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>{created ? "Invitation created" : "Invite member"}</DialogTitle>
          <DialogDescription>
            {created
              ? created.emailed
                ? `We emailed ${created.email} a link to join. You can also share the link yourself.`
                : `Email is not set up, so share this link with ${created.email}. It works once.`
              : "Invite someone by email. They join with the role and teams you pick."}
          </DialogDescription>
        </DialogHeader>
        {isOpen && !created && <InviteForm onCreated={setCreated} />}
        {created && (
          <>
            {created.link && <CopyField value={created.link} label="Invitation link" />}
            <DialogFooter>
              <Button size="sm" variant="outline" onClick={() => setCreated(null)}>
                Invite another
              </Button>
              <DialogClose render={<Button size="sm">Done</Button>} />
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

function InviteForm({ onCreated }: { onCreated: (invitation: TCreatedInvitation) => void }) {
  const { workspace } = useWorkspace();
  const { teams } = useTeams();
  const remember = useInvitationLinksStore((state) => state.remember);
  const { mutateAsync: createInvitation } = useCreateInvitationMutation(workspace.id);
  const form = useAppForm({
    defaultValues: { email: "", role: "member" as TWorkspaceRole, team_ids: [] as string[] },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      const res = await createInvitation({ workspace_id: workspace.id, ...parsed });
      if (res.data.link) remember(res.data.id, res.data.link);
      if (res.data.emailed) {
        toast.add({ type: "success", title: `Invitation sent to ${res.data.email}` });
      }
      onCreated(res.data);
    },
  });

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="email"
        children={(field) => (
          <field.TextField label="Email" type="email" placeholder="name@example.com" autoFocus />
        )}
      />
      <form.AppField
        name="role"
        children={(field) => (
          <field.SelectField
            label="Role"
            description="Viewers read, members edit, admins also manage the workspace."
            items={roleSelectItems}
          />
        )}
      />
      <form.Field
        name="team_ids"
        children={(field) => (
          <div className="flex flex-col gap-1.5">
            <p className="text-sm font-medium">Teams</p>
            <CheckboxList
              label="Teams"
              emptyText="There are no teams yet."
              items={teams.map((team) => ({
                value: team.id,
                label: team.name,
                description: team.key,
                icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
              }))}
              value={field.state.value}
              onChange={field.handleChange}
            />
          </div>
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
          <form.SubmitButton size="sm">Send invitation</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
