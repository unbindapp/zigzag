import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, redirect, useRouter } from "@tanstack/react-router";
import { z } from "zod";

import { AuthShell } from "@/components/layout/auth-shell";
import { slugify, workspaceSlugPattern } from "@/lib/helpers/identifiers";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { meQuery } from "@/lib/queries/me";
import { createBootstrapUser, queryKeySetup, setupStatusQuery } from "@/lib/queries/setup";

const formSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  display_name: z
    .string()
    .trim()
    .regex(/^[a-z0-9._-]{2,32}$/, "Use 2 to 32 lowercase letters, numbers, dots or dashes."),
  email: z.string().trim().email("Enter a valid email address."),
  password: z.string().min(8, "Password must be at least 8 characters."),
  workspace_name: z.string().trim().min(1, "Workspace name is required."),
  workspace_slug: z
    .string()
    .trim()
    .regex(workspaceSlugPattern, "Use 2 to 32 lowercase letters, numbers or dashes."),
});

export const Route = createFileRoute("/welcome/")({
  beforeLoad: async ({ context }) => {
    const { status } = await context.queryClient.fetchQuery(setupStatusQuery());
    if (status.is_first_user_created) throw redirect({ to: "/sign-in" });
  },
  component: Welcome,
});

function handleFromName(name: string) {
  return slugify(name.split(/\s+/)[0] ?? "", 32).replace(/-/g, "");
}

function Welcome() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { mutateAsync: bootstrap } = useMutation({
    mutationFn: createBootstrapUser,
    meta: { errorTitle: "Could not finish setup" },
  });

  const form = useAppForm({
    defaultValues: {
      name: "",
      display_name: "",
      email: "",
      password: "",
      workspace_name: "",
      workspace_slug: "",
    },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      const res = await bootstrap(parsed);
      queryClient.removeQueries({ queryKey: queryKeySetup.status() });
      queryClient.removeQueries({ queryKey: meQuery.queryKey });
      await queryClient.ensureQueryData(meQuery);
      router.history.push(`/${res.data.workspace.slug}`);
    },
  });

  return (
    <AuthShell
      title="Welcome to Zigzag"
      description="Create the first account and workspace. This account becomes the system admin."
    >
      <form className="flex w-full flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
        <form.AppField
          name="name"
          listeners={{
            onChange: ({ value }) => {
              if (form.getFieldMeta("display_name")?.isDirty) return;
              form.setFieldValue("display_name", handleFromName(value), { dontUpdateMeta: true });
            },
          }}
          children={(field) => <field.TextField label="Full name" autoComplete="name" autoFocus />}
        />
        <form.AppField
          name="display_name"
          children={(field) => (
            <field.TextField
              label="Username"
              description="Used for @mentions."
              autoComplete="username"
              transform={(v) => v.toLowerCase()}
            />
          )}
        />
        <form.AppField
          name="email"
          children={(field) => <field.TextField label="Email" type="email" autoComplete="email" />}
        />
        <form.AppField
          name="password"
          children={(field) => (
            <field.TextField label="Password" type="password" autoComplete="new-password" />
          )}
        />
        <form.AppField
          name="workspace_name"
          listeners={{
            onChange: ({ value }) => {
              if (form.getFieldMeta("workspace_slug")?.isDirty) return;
              form.setFieldValue("workspace_slug", slugify(value), { dontUpdateMeta: true });
            },
          }}
          children={(field) => <field.TextField label="Workspace name" placeholder="Acme" />}
        />
        <form.AppField
          name="workspace_slug"
          children={(field) => (
            <field.TextField
              label="Workspace URL"
              description="Lowercase letters, numbers and dashes."
              transform={(v) => v.toLowerCase()}
            />
          )}
        />
        <form.AppForm>
          <form.SubmitButton className="w-full">Create account and workspace</form.SubmitButton>
        </form.AppForm>
      </form>
    </AuthShell>
  );
}
