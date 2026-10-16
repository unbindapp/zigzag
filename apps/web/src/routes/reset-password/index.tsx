import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { CircleCheckIcon, LinkIcon } from "lucide-react";
import { z } from "zod";

import { AuthShell } from "@/components/layout/auth-shell";
import { EmptyState } from "@/components/layout/empty-state";
import { LinkButton } from "@/components/ui/button";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { resetPassword } from "@/lib/queries/auth";

const searchSchema = z.object({ token: z.string().optional() });

const formSchema = z
  .object({
    password: z.string().min(8, "Password must be at least 8 characters."),
    confirm: z.string(),
  })
  .refine((value) => value.password === value.confirm, {
    message: "Passwords do not match.",
    path: ["confirm"],
  });

export const Route = createFileRoute("/reset-password/")({
  validateSearch: zodValidator(searchSchema),
  component: ResetPassword,
});

function ResetPassword() {
  const { token } = Route.useSearch();
  const { mutateAsync: reset, isSuccess } = useMutation({
    mutationFn: resetPassword,
    meta: { errorTitle: "Could not reset password" },
  });

  const form = useAppForm({
    defaultValues: { password: "", confirm: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      if (!token) return;
      await reset({ token, password: value.password });
    },
  });

  if (!token) {
    return (
      <AuthShell title="Reset your password">
        <EmptyState
          icon={LinkIcon}
          title="This link is incomplete"
          description="Open the link from the email again, or request a new one."
          action={
            <LinkButton to="/forgot-password" variant="outline">
              Request a new link
            </LinkButton>
          }
        />
      </AuthShell>
    );
  }

  if (isSuccess) {
    return (
      <AuthShell title="Password updated">
        <EmptyState
          icon={CircleCheckIcon}
          title="You can sign in with your new password"
          action={<LinkButton to="/sign-in">Sign in</LinkButton>}
        />
      </AuthShell>
    );
  }

  return (
    <AuthShell title="Choose a new password">
      <form className="flex w-full flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
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
        <form.AppField
          name="confirm"
          children={(field) => (
            <field.TextField label="Confirm password" type="password" autoComplete="new-password" />
          )}
        />
        <form.AppForm>
          <form.SubmitButton className="w-full">Set password</form.SubmitButton>
        </form.AppForm>
      </form>
    </AuthShell>
  );
}
