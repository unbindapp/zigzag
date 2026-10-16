import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { toast } from "@/components/ui/toast";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useChangePasswordMutation } from "@/lib/queries/me";

export const Route = createFileRoute("/$workspace/_workspace/settings/account/password")({
  component: AccountPasswordSettingsPage,
});

const formSchema = z
  .object({
    current: z.string().min(1, "Enter your current password."),
    next: z.string().min(8, "Password must be at least 8 characters."),
    confirm: z.string(),
  })
  .refine((value) => value.next === value.confirm, {
    message: "Passwords do not match.",
    path: ["confirm"],
  });

function AccountPasswordSettingsPage() {
  const { mutateAsync: changePassword } = useChangePasswordMutation();
  const form = useAppForm({
    defaultValues: { current: "", next: "", confirm: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value, formApi }) => {
      await changePassword({ currentPassword: value.current, newPassword: value.next });
      formApi.reset();
      toast.add({ type: "success", title: "Password changed" });
    },
  });
  return (
    <SettingsPage title="Password" description="Change the password you sign in with.">
      <SettingsSection
        title="Change password"
        onSubmit={submitHandler(form)}
        footer={
          <form.AppForm>
            <form.SubmitButton size="sm">Change password</form.SubmitButton>
          </form.AppForm>
        }
      >
        <form.AppField
          name="current"
          children={(field) => (
            <field.TextField
              label="Current password"
              type="password"
              autoComplete="current-password"
            />
          )}
        />
        <form.AppField
          name="next"
          children={(field) => (
            <field.TextField label="New password" type="password" autoComplete="new-password" />
          )}
        />
        <form.AppField
          name="confirm"
          children={(field) => (
            <field.TextField
              label="Confirm new password"
              type="password"
              autoComplete="new-password"
            />
          )}
        />
      </SettingsSection>
    </SettingsPage>
  );
}
