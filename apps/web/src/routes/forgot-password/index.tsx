import { useMutation } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { MailCheckIcon } from "lucide-react";
import { z } from "zod";

import { AuthShell } from "@/components/layout/auth-shell";
import { EmptyState } from "@/components/layout/empty-state";
import { LinkButton } from "@/components/ui/button";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { requestPasswordReset } from "@/lib/queries/auth";

const formSchema = z.object({
  email: z.string().trim().email("Enter a valid email address."),
});

export const Route = createFileRoute("/forgot-password/")({
  component: ForgotPassword,
});

function ForgotPassword() {
  const { mutateAsync: request, isSuccess } = useMutation({
    mutationFn: requestPasswordReset,
    meta: { errorTitle: "Could not send reset link" },
  });

  const form = useAppForm({
    defaultValues: { email: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      await request({ email: value.email.trim() });
    },
  });

  if (isSuccess) {
    return (
      <AuthShell title="Check your email">
        <EmptyState
          icon={MailCheckIcon}
          title="Reset link sent"
          description="If an account exists for that address, it will receive a link to set a new password. Ask an admin to reset it if email is not set up."
          action={
            <LinkButton to="/sign-in" variant="outline">
              Back to sign in
            </LinkButton>
          }
        />
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title="Reset your password"
      description="Enter your email and we will send you a reset link."
      footer={
        <Link
          to="/sign-in"
          className="has-hover:hover:text-foreground underline-offset-4 has-hover:hover:underline"
        >
          Back to sign in
        </Link>
      }
    >
      <form className="flex w-full flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
        <form.AppField
          name="email"
          children={(field) => (
            <field.TextField label="Email" type="email" autoComplete="email" autoFocus />
          )}
        />
        <form.AppForm>
          <form.SubmitButton className="w-full">Send reset link</form.SubmitButton>
        </form.AppForm>
      </form>
    </AuthShell>
  );
}
