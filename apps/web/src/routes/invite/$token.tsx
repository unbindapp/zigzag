import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { MailXIcon } from "lucide-react";
import { z } from "zod";

import { AuthShell } from "@/components/layout/auth-shell";
import { EmptyState } from "@/components/layout/empty-state";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { Button, LinkButton } from "@/components/ui/button";
import { formatDate } from "@/lib/helpers/dates";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  invitationPreviewQuery,
  type TInvitationPreview,
  useAcceptInvitationMutation,
} from "@/lib/queries/invitations";
import { meQuery } from "@/lib/queries/me";
import { workspaceRoleLabels } from "@/lib/queries/workspaces";

export const Route = createFileRoute("/invite/$token")({
  component: Invite,
});

function Invite() {
  const { token } = Route.useParams();
  const { data, isPending, error } = useQuery(invitationPreviewQuery(token));

  if (isPending) {
    return (
      <AuthShell title="Loading your invitation">
        <ListSkeleton rows={3} />
      </AuthShell>
    );
  }

  if (!data) {
    return (
      <AuthShell title="This invitation is no longer valid">
        <EmptyState
          icon={MailXIcon}
          title="Invitation not found"
          description={error?.message ?? "Ask the person who invited you for a new link."}
          action={
            <LinkButton to="/sign-in" variant="outline">
              Go to sign in
            </LinkButton>
          }
        />
      </AuthShell>
    );
  }

  return <AcceptInvitation token={token} invitation={data.invitation} />;
}

function AcceptInvitation({
  token,
  invitation,
}: {
  token: string;
  invitation: TInvitationPreview;
}) {
  const { data: me } = useQuery(meQuery);
  const title = `Join ${invitation.workspace_name}`;
  const description = `You were invited as ${workspaceRoleLabels[invitation.role].toLowerCase()} with ${invitation.email}. The invitation expires on ${formatDate(invitation.expires_at)}.`;

  if (!invitation.account_exists) {
    return (
      <AuthShell title={title} description={description}>
        <NewAccountForm token={token} />
      </AuthShell>
    );
  }

  if (!me) {
    return (
      <AuthShell title={title} description={description}>
        <LinkButton to="/sign-in" search={{ redirect: `/invite/${token}` }} className="w-full">
          Sign in to accept
        </LinkButton>
      </AuthShell>
    );
  }

  return (
    <AuthShell title={title} description={description}>
      <AcceptButton token={token} />
    </AuthShell>
  );
}

function useAfterAccept() {
  const router = useRouter();
  const queryClient = useQueryClient();
  return async (slug: string) => {
    await queryClient.ensureQueryData(meQuery);
    router.history.push(`/${slug}`);
  };
}

function AcceptButton({ token }: { token: string }) {
  const { mutateAsync: accept, isPending } = useAcceptInvitationMutation();
  const afterAccept = useAfterAccept();
  return (
    <Button
      className="w-full"
      isPending={isPending}
      onClick={async () => {
        const res = await accept({ token }).catch(() => null);
        if (!res) return;
        await afterAccept(res.data.workspace.slug);
      }}
    >
      Accept invitation
    </Button>
  );
}

const newAccountSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  display_name: z
    .string()
    .trim()
    .regex(/^[a-z0-9._-]{2,32}$/, "Use 2 to 32 lowercase letters, numbers, dots or dashes."),
  password: z.string().min(8, "Password must be at least 8 characters."),
});

function NewAccountForm({ token }: { token: string }) {
  const { mutateAsync: accept } = useAcceptInvitationMutation();
  const afterAccept = useAfterAccept();
  const form = useAppForm({
    defaultValues: { name: "", display_name: "", password: "" },
    validators: { onSubmit: newAccountSchema },
    onSubmit: async ({ value }) => {
      const res = await accept({ token, ...newAccountSchema.parse(value) });
      await afterAccept(res.data.workspace.slug);
    },
  });
  return (
    <form className="flex w-full flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="name"
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
        name="password"
        children={(field) => (
          <field.TextField label="Password" type="password" autoComplete="new-password" />
        )}
      />
      <form.AppForm>
        <form.SubmitButton className="w-full">Create account and join</form.SubmitButton>
      </form.AppForm>
    </form>
  );
}
