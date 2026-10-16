import { useQuery } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";

import { AuthShell } from "@/components/layout/auth-shell";
import { Button } from "@/components/ui/button";
import { slugify } from "@/lib/helpers/identifiers";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useLogoutMutation } from "@/lib/queries/auth";
import { meQuery, meWorkspaces } from "@/lib/queries/me";
import { useCreateWorkspaceMutation, WorkspaceFormSchema } from "@/lib/queries/workspaces";

export const Route = createFileRoute("/create-workspace/")({
  component: CreateWorkspace,
});

function CreateWorkspace() {
  const router = useRouter();
  const { data: me } = useQuery(meQuery);
  const hasWorkspaces = meWorkspaces(me).length > 0;
  const { mutateAsync: create } = useCreateWorkspaceMutation();
  const { mutate: signOut, isPending: isSigningOut } = useLogoutMutation();

  const form = useAppForm({
    defaultValues: { name: "", slug: "" },
    validators: { onSubmit: WorkspaceFormSchema },
    onSubmit: async ({ value }) => {
      const parsed = WorkspaceFormSchema.parse(value);
      const res = await create(parsed);
      router.history.push(`/${res.data.slug}`);
    },
  });

  return (
    <AuthShell
      title="Create a workspace"
      description={
        hasWorkspaces
          ? "Workspaces hold teams, projects and issues."
          : "You are not a member of any workspace yet. Create one, or ask an admin for an invitation."
      }
      footer={
        <Button
          variant="ghost"
          size="sm"
          isPending={isSigningOut}
          onClick={() => signOut(undefined, { onSuccess: () => router.history.push("/sign-in") })}
        >
          Sign out
        </Button>
      }
    >
      <form className="flex w-full flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
        <form.AppField
          name="name"
          listeners={{
            onChange: ({ value }) => {
              if (form.getFieldMeta("slug")?.isDirty) return;
              form.setFieldValue("slug", slugify(value), { dontUpdateMeta: true });
            },
          }}
          children={(field) => <field.TextField label="Name" placeholder="Acme" autoFocus />}
        />
        <form.AppField
          name="slug"
          children={(field) => (
            <field.TextField
              label="URL"
              description="Lowercase letters, numbers and dashes."
              transform={(v) => v.toLowerCase()}
            />
          )}
        />
        <form.AppForm>
          <form.SubmitButton className="w-full">Create workspace</form.SubmitButton>
        </form.AppForm>
      </form>
    </AuthShell>
  );
}
