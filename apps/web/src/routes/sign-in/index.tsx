import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link, redirect, useRouter } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { useMutation } from "@tanstack/react-query";
import { z } from "zod";

import { GoogleIcon } from "@/components/icons/google";
import { AuthShell } from "@/components/layout/auth-shell";
import { ErrorCard } from "@/components/layout/error-card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { login } from "@/lib/queries/auth";
import { authProvidersQuery, googleSignInHref } from "@/lib/queries/google-auth";
import { meQuery } from "@/lib/queries/me";
import { setupStatusQuery } from "@/lib/queries/setup";

const searchSchema = z.object({
  redirect: z.string().optional(),
  error: z.string().optional(),
});

const signInErrors: Record<string, string> = {
  no_account: "No account for this Google email. Ask an admin for an invitation.",
  google_failed: "Google sign-in did not work. Try again or sign in with your password.",
};

const formSchema = z.object({
  email: z.string().trim().email("Enter a valid email address."),
  password: z.string().min(1, "Password is required."),
});

export const Route = createFileRoute("/sign-in/")({
  validateSearch: zodValidator(searchSchema),
  beforeLoad: async ({ context }) => {
    const me = await context.queryClient.ensureQueryData(meQuery);
    if (me) throw redirect({ to: "/" });
    const isFirstUserCreated = await context.queryClient
      .fetchQuery(setupStatusQuery())
      .then((res) => res.status.is_first_user_created)
      .catch(() => true);
    if (!isFirstUserCreated) throw redirect({ to: "/welcome" });
  },
  component: SignIn,
});

function SignIn() {
  const router = useRouter();
  const search = Route.useSearch();
  const queryClient = useQueryClient();
  const { data: providers } = useQuery(authProvidersQuery());
  const errorMessage = search.error ? signInErrors[search.error] : undefined;
  const { mutateAsync: signIn } = useMutation({
    mutationFn: login,
    meta: { errorTitle: "Could not sign in" },
  });

  const form = useAppForm({
    defaultValues: { email: "", password: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      await signIn({ email: value.email.trim(), password: value.password });
      queryClient.removeQueries({ queryKey: meQuery.queryKey });
      await queryClient.ensureQueryData(meQuery);
      router.history.push(search.redirect ?? "/");
    },
  });

  return (
    <AuthShell
      title="Sign in to Zigzag"
      footer={
        <Link
          to="/forgot-password"
          className="has-hover:hover:text-foreground underline-offset-4 has-hover:hover:underline"
        >
          Forgot password?
        </Link>
      }
    >
      {errorMessage && <ErrorCard title={errorMessage} />}
      <div className="flex w-full flex-col gap-4">
        {providers?.providers.google.enabled && (
          <>
            <Button
              variant="outline"
              className="w-full"
              render={<a href={googleSignInHref(search.redirect)} />}
            >
              <GoogleIcon />
              Continue with Google
            </Button>
            <div className="flex items-center gap-2">
              <Separator className="flex-1" />
              <span className="text-muted-foreground text-xs">or</span>
              <Separator className="flex-1" />
            </div>
          </>
        )}
        <form className="flex w-full flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
          <form.AppField
            name="email"
            children={(field) => (
              <field.TextField label="Email" type="email" autoComplete="email" autoFocus />
            )}
          />
          <form.AppField
            name="password"
            children={(field) => (
              <field.TextField label="Password" type="password" autoComplete="current-password" />
            )}
          />
          <form.AppForm>
            <form.SubmitButton className="w-full">Sign in</form.SubmitButton>
          </form.AppForm>
        </form>
      </div>
    </AuthShell>
  );
}
