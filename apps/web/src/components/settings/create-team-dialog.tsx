import { z } from "zod";

import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useCreateTeamStore } from "@/components/settings/create-team-store";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { deriveTeamKey, sanitizeTeamKey } from "@/lib/helpers/identifiers";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { TeamFormSchema, useCreateTeamMutation } from "@/lib/queries/teams";

const formSchema = TeamFormSchema.extend({
  icon: z.string(),
  color: z.string(),
  private: z.boolean(),
});

export function CreateTeamDialog() {
  const isOpen = useCreateTeamStore((state) => state.isOpen);
  const setOpen = useCreateTeamStore((state) => state.setOpen);
  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      <DialogContent className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>Create team</DialogTitle>
          <DialogDescription>
            The key prefixes every issue identifier, like ENG-123.
          </DialogDescription>
        </DialogHeader>
        {isOpen && <CreateTeamForm onDone={() => setOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function CreateTeamForm({ onDone }: { onDone: () => void }) {
  const { workspace } = useWorkspace();
  const { mutateAsync: createTeam } = useCreateTeamMutation();
  const form = useAppForm({
    defaultValues: { name: "", key: "", icon: "", color: "", private: false },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      await createTeam({
        workspace_id: workspace.id,
        name: parsed.name,
        key: parsed.key,
        icon: value.icon || undefined,
        color: value.color || undefined,
        private: value.private,
      });
      onDone();
    },
  });

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <div className="flex items-start gap-4">
        <form.Subscribe selector={(state) => state.values.color}>
          {(color) => (
            <form.AppField
              name="icon"
              children={(field) => <field.IconField label="Icon" color={color} />}
            />
          )}
        </form.Subscribe>
        <form.AppField name="color" children={(field) => <field.ColorField label="Color" />} />
      </div>
      <form.AppField
        name="name"
        listeners={{
          onChange: ({ value }) => {
            if (form.getFieldMeta("key")?.isDirty) return;
            form.setFieldValue("key", deriveTeamKey(value), { dontUpdateMeta: true });
          },
        }}
        children={(field) => <field.TextField label="Name" placeholder="Engineering" autoFocus />}
      />
      <form.AppField
        name="key"
        children={(field) => (
          <field.TextField
            label="Key"
            description="Up to 7 letters or digits, starting with a letter."
            transform={sanitizeTeamKey}
            className="font-mono"
          />
        )}
      />
      <form.AppField
        name="private"
        children={(field) => (
          <field.SwitchField
            label="Private team"
            description="Only members and workspace admins can see a private team."
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
          <form.SubmitButton size="sm">Create team</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
