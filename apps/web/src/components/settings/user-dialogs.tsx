import { ReactElement, useState } from "react";
import { z } from "zod";

import { githubLoginSchema } from "@/components/integrations/github-sync";
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
import { type TUser, useCreateUserMutation, useUpdateUserMutation } from "@/lib/queries/users";

const handleSchema = z
  .string()
  .trim()
  .regex(/^[a-z0-9._-]{2,32}$/, "Use 2 to 32 lowercase letters, numbers, dots or dashes.");

const createSchema = z.object({
  email: z.string().trim().email("Enter a valid email address."),
  name: z.string().trim().min(1, "Name is required."),
  display_name: handleSchema,
  password: z.string().min(8, "Password must be at least 8 characters."),
  system_admin: z.boolean(),
});

const editSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  display_name: handleSchema,
  github_login: githubLoginSchema,
  system_admin: z.boolean(),
  active: z.boolean(),
});

const passwordSchema = z.object({
  password: z.string().min(8, "Password must be at least 8 characters."),
});

function FormDialog({
  trigger,
  title,
  description,
  children,
}: {
  trigger: ReactElement;
  title: string;
  description?: string;
  children: (close: () => void) => ReactElement;
}) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger render={trigger} />
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        {isOpen && children(() => setIsOpen(false))}
      </DialogContent>
    </Dialog>
  );
}

export function CreateUserDialog({ trigger }: { trigger: ReactElement }) {
  return (
    <FormDialog
      trigger={trigger}
      title="Create user"
      description="The user can sign in right away with this password."
    >
      {(close) => <CreateUserForm onDone={close} />}
    </FormDialog>
  );
}

function CreateUserForm({ onDone }: { onDone: () => void }) {
  const { mutateAsync: createUser } = useCreateUserMutation();
  const form = useAppForm({
    defaultValues: { email: "", name: "", display_name: "", password: "", system_admin: false },
    validators: { onSubmit: createSchema },
    onSubmit: async ({ value }) => {
      await createUser(createSchema.parse(value));
      onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="email"
        children={(field) => <field.TextField label="Email" type="email" autoFocus />}
      />
      <form.AppField name="name" children={(field) => <field.TextField label="Full name" />} />
      <form.AppField
        name="display_name"
        children={(field) => (
          <field.TextField label="Username" transform={(v) => v.toLowerCase()} />
        )}
      />
      <form.AppField
        name="password"
        children={(field) => (
          <field.TextField label="Password" type="password" autoComplete="new-password" />
        )}
      />
      <form.AppField
        name="system_admin"
        children={(field) => (
          <field.SwitchField
            label="System admin"
            description="System admins manage instance settings and users, and are admins in every workspace."
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
          <form.SubmitButton size="sm">Create user</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}

export function EditUserDialog({
  user,
  trigger,
  isSelf,
}: {
  user: TUser;
  trigger: ReactElement;
  isSelf: boolean;
}) {
  return (
    <FormDialog trigger={trigger} title={`Edit ${user.name || user.email}`}>
      {(close) => <EditUserForm user={user} isSelf={isSelf} onDone={close} />}
    </FormDialog>
  );
}

function EditUserForm({
  user,
  isSelf,
  onDone,
}: {
  user: TUser;
  isSelf: boolean;
  onDone: () => void;
}) {
  const { mutateAsync: updateUser } = useUpdateUserMutation();
  const form = useAppForm({
    defaultValues: {
      name: user.name,
      display_name: user.display_name,
      github_login: user.github_login ?? "",
      system_admin: user.system_admin,
      active: user.active,
    },
    validators: { onSubmit: editSchema },
    onSubmit: async ({ value }) => {
      await updateUser({ user_id: user.id, ...editSchema.parse(value) });
      onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField name="name" children={(field) => <field.TextField label="Full name" />} />
      <form.AppField
        name="display_name"
        children={(field) => (
          <field.TextField label="Username" transform={(v) => v.toLowerCase()} />
        )}
      />
      <form.AppField
        name="github_login"
        children={(field) => (
          <field.TextField
            label="GitHub username"
            placeholder="octocat"
            autoComplete="off"
            description="Used by GitHub Issues sync to credit and assign this person."
          />
        )}
      />
      <form.AppField
        name="system_admin"
        children={(field) => (
          <field.SwitchField
            label="System admin"
            description={isSelf ? "You cannot remove your own system admin role." : undefined}
            disabled={isSelf}
          />
        )}
      />
      <form.AppField
        name="active"
        children={(field) => (
          <field.SwitchField
            label="Active"
            description={
              isSelf ? "You cannot deactivate yourself." : "Inactive users cannot sign in."
            }
            disabled={isSelf}
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
          <form.SubmitButton size="sm" requireDirty>
            Save
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}

export function ResetPasswordDialog({ user, trigger }: { user: TUser; trigger: ReactElement }) {
  return (
    <FormDialog
      trigger={trigger}
      title="Reset password"
      description={`Set a new password for ${user.email}. Share it with them securely.`}
    >
      {(close) => <ResetPasswordForm user={user} onDone={close} />}
    </FormDialog>
  );
}

function ResetPasswordForm({ user, onDone }: { user: TUser; onDone: () => void }) {
  const { mutateAsync: updateUser } = useUpdateUserMutation();
  const form = useAppForm({
    defaultValues: { password: "" },
    validators: { onSubmit: passwordSchema },
    onSubmit: async ({ value }) => {
      await updateUser({ user_id: user.id, password: value.password });
      onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="password"
        children={(field) => (
          <field.TextField
            label="New password"
            type="password"
            autoComplete="new-password"
            autoFocus
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
          <form.SubmitButton size="sm">Set password</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
