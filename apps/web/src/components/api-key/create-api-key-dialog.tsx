import { PlusIcon } from "lucide-react";
import { useState } from "react";
import { z } from "zod";

import { roleDescription, WorkspaceAccessList } from "@/components/api-key/access-fields";
import { ApiKeyCreatedDialog } from "@/components/api-key/api-key-created-dialog";
import {
  accessFormShape,
  accessItems,
  defaultExpiry,
  expiresAtFrom,
  expiryOptions,
  hasPickedWorkspace,
  pickWorkspaceMessage,
  type TAccess,
  type TExpiryValue,
} from "@/components/api-key/helpers";
import { roleSelectItems } from "@/components/shared/role-select";
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
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { type TApiKeyCreated, useCreateApiKeyMutation } from "@/lib/queries/api-keys";
import type { TWorkspaceRole } from "@/lib/queries/workspaces";

const formSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, "Name is required.")
      .max(100, "Keep the name under 100 characters."),
    ...accessFormShape,
    expiry: z.enum(["7d", "30d", "90d", "1y", "never"]),
  })
  .refine(hasPickedWorkspace, pickWorkspaceMessage);

export function CreateApiKeyDialog({ takenNames }: { takenNames: string[] }) {
  const [isOpen, setIsOpen] = useState(false);
  const [created, setCreated] = useState<TApiKeyCreated | null>(null);
  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger
          render={
            <Button size="sm">
              <PlusIcon className="size-4" />
              Create key
            </Button>
          }
        />
        <DialogContent className="w-lg max-w-full">
          <DialogHeader>
            <DialogTitle>Create API key</DialogTitle>
            <DialogDescription>
              API keys let scripts and agents use Zigzag as you, limited to the role and workspaces
              you pick.
            </DialogDescription>
          </DialogHeader>
          {isOpen && (
            <CreateApiKeyForm
              takenNames={takenNames}
              onCreated={(key) => {
                setIsOpen(false);
                setCreated(key);
              }}
            />
          )}
        </DialogContent>
      </Dialog>
      <ApiKeyCreatedDialog created={created} onClose={() => setCreated(null)} />
    </>
  );
}

function CreateApiKeyForm({
  takenNames,
  onCreated,
}: {
  takenNames: string[];
  onCreated: (key: TApiKeyCreated) => void;
}) {
  const { mutateAsync: createKey } = useCreateApiKeyMutation();
  const form = useAppForm({
    defaultValues: {
      name: "",
      role: "member" as TWorkspaceRole,
      access: "all" as TAccess,
      workspace_ids: [] as string[],
      expiry: defaultExpiry as TExpiryValue,
    },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const res = await createKey({
        name: value.name.trim(),
        role: value.role,
        full_access: value.access === "all",
        workspace_ids: value.access === "all" ? [] : value.workspace_ids,
        expires_at: expiresAtFrom(value.expiry),
      });
      onCreated(res.data);
    },
  });

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="name"
        validators={{
          onSubmit: ({ value }) =>
            takenNames.includes(value.trim())
              ? "You already have a key with this name."
              : undefined,
        }}
        children={(field) => (
          <field.TextField
            label="Name"
            description="Something that tells you where the key is used."
            placeholder="CI pipeline"
            autoFocus
          />
        )}
      />
      <form.AppField
        name="role"
        children={(field) => (
          <field.SelectField
            label="Role"
            description={roleDescription("key")}
            items={roleSelectItems}
          />
        )}
      />
      <form.AppField
        name="access"
        children={(field) => <field.SelectField label="Access" items={accessItems} />}
      />
      <form.Subscribe selector={(state) => state.values.access}>
        {(access) => (
          <form.Field
            name="workspace_ids"
            children={(field) => <WorkspaceAccessList field={field} access={access} />}
          />
        )}
      </form.Subscribe>
      <form.AppField
        name="expiry"
        children={(field) => (
          <field.SelectField
            label="Expires"
            description="An expired key stops working immediately."
            items={expiryOptions.map((option) => ({ value: option.value, label: option.label }))}
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
          <form.SubmitButton size="sm">Create key</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
