import { useStore } from "@tanstack/react-form";
import { useNavigate } from "@tanstack/react-router";
import { CircleUserIcon } from "lucide-react";
import { type ReactElement, useState } from "react";

import { userName } from "@/components/initiatives/initiative-pickers";
import {
  INITIATIVE_STATUSES,
  InitiativeStatusIcon,
  initiativeStatusLabels,
} from "@/components/initiatives/initiative-status";
import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { Avatar } from "@/components/shared/avatar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, type TSelectFieldItem, useAppForm } from "@/lib/hooks/use-app-form";
import {
  InitiativeFormSchema,
  type TInitiativeStatus,
  useCreateInitiativeMutation,
} from "@/lib/queries/initiatives";

const NO_OWNER = "__none__";

const statusItems: TSelectFieldItem[] = INITIATIVE_STATUSES.map((status) => ({
  value: status,
  label: initiativeStatusLabels[status],
  icon: <InitiativeStatusIcon status={status} />,
}));

export function NewInitiativeDialog({ trigger }: { trigger: ReactElement }) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger render={trigger} />
      <DialogContent className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>New initiative</DialogTitle>
        </DialogHeader>
        {isOpen && <NewInitiativeForm onDone={() => setIsOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function NewInitiativeForm({ onDone }: { onDone: () => void }) {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { users } = useWorkspaceUsers();
  const { mutateAsync: createInitiative } = useCreateInitiativeMutation();
  const ownerItems: TSelectFieldItem[] = [
    {
      value: NO_OWNER,
      label: "No owner",
      icon: <CircleUserIcon className="text-muted-foreground size-4" />,
    },
    ...users.map((user) => ({
      value: user.id,
      label: userName(user),
      icon: <Avatar id={user.id} name={userName(user)} src={user.avatar_url} size="xs" />,
    })),
  ];

  const form = useAppForm({
    defaultValues: {
      name: "",
      description: "",
      status: "planned" as TInitiativeStatus,
      owner_id: NO_OWNER,
      target_date: "",
      icon: "",
      color: "",
    },
    validators: { onSubmit: InitiativeFormSchema },
    onSubmit: async ({ value }) => {
      const parsed = InitiativeFormSchema.parse(value);
      const res = await createInitiative({
        workspace_id: workspace.id,
        name: parsed.name,
        description: parsed.description || undefined,
        status: parsed.status,
        owner_id: parsed.owner_id === NO_OWNER ? undefined : parsed.owner_id,
        target_date: parsed.target_date || undefined,
        icon: parsed.icon || undefined,
        color: parsed.color || undefined,
      });
      onDone();
      await navigate({
        to: "/$workspace/initiative/$slug",
        params: { workspace: workspace.slug, slug: res.data.slug },
      });
    },
  });
  const color = useStore(form.store, (state) => state.values.color);

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <div className="flex items-start gap-4">
        <form.AppField
          name="icon"
          children={(field) => <field.IconField label="Icon" color={color || undefined} />}
        />
        <form.AppField name="color" children={(field) => <field.ColorField label="Color" />} />
      </div>
      <form.AppField name="name" children={(field) => <field.TextField label="Name" autoFocus />} />
      <form.AppField
        name="description"
        children={(field) => (
          <field.TextareaField
            label="Description"
            rows={3}
            placeholder="What is this initiative about?"
          />
        )}
      />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <form.AppField
          name="status"
          children={(field) => <field.SelectField label="Status" items={statusItems} />}
        />
        <form.AppField
          name="owner_id"
          children={(field) => <field.SelectField label="Owner" items={ownerItems} />}
        />
      </div>
      <form.AppField
        name="target_date"
        children={(field) => <field.TextField label="Target date" type="date" />}
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
          <form.SubmitButton size="sm">Create initiative</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
