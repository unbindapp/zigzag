import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { UploadIcon, XIcon } from "lucide-react";
import { z } from "zod";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { Avatar } from "@/components/shared/avatar";
import { Button } from "@/components/ui/button";
import { FileTrigger } from "@/components/ui/file-trigger";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { timezoneOptions } from "@/lib/helpers/timezones";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { meQuery, NO_AVATAR_UPLOAD_ID, type TMe, useUpdateMeMutation } from "@/lib/queries/me";
import { useCreateUploadMutation } from "@/lib/queries/uploads";

export const Route = createFileRoute("/$workspace/_workspace/settings/account/")({
  component: AccountProfileSettingsPage,
});

const timezones = timezoneOptions();

const formSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  display_name: z
    .string()
    .trim()
    .regex(/^[a-z0-9._-]{2,32}$/, "Use 2 to 32 lowercase letters, numbers, dots or dashes."),
  timezone: z.string().min(1),
});

function AccountProfileSettingsPage() {
  const { data: me, isPending, error, refetch } = useQuery(meQuery);
  return (
    <SettingsPage title="Profile" description="How you appear to others in Zigzag.">
      {isPending && <ListSkeleton rows={3} />}
      {!isPending && error && <ErrorCard message={error.message} onRetry={() => void refetch()} />}
      {me && (
        <>
          <AvatarSection me={me} />
          <ProfileSection me={me} />
        </>
      )}
    </SettingsPage>
  );
}

function AvatarSection({ me }: { me: TMe }) {
  const { workspace } = useWorkspace();
  const { mutateAsync: upload, isPending: isUploading } = useCreateUploadMutation();
  const { mutateAsync: updateMe, isPending: isSaving } = useUpdateMeMutation();
  const user = me.user;
  return (
    <SettingsSection title="Avatar" description="A square image works best.">
      <div className="flex items-center gap-4">
        <Avatar id={user.id} name={user.name} src={user.avatar_url} size="xl" />
        <FileTrigger
          accept="image/*"
          onSelect={async ([file]) => {
            const res = await upload({ workspaceId: workspace.id, file }).catch(() => null);
            if (!res) return;
            await updateMe({ avatar_upload_id: res.data.id }).catch(() => undefined);
          }}
        >
          {(open) => (
            <Button variant="outline" size="sm" onClick={open} isPending={isUploading || isSaving}>
              <UploadIcon className="size-4" />
              Upload image
            </Button>
          )}
        </FileTrigger>
        {user.avatar_url && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() =>
              updateMe({ avatar_upload_id: NO_AVATAR_UPLOAD_ID }).catch(() => undefined)
            }
          >
            <XIcon className="size-4" />
            Remove
          </Button>
        )}
      </div>
    </SettingsSection>
  );
}

function ProfileSection({ me }: { me: TMe }) {
  const { mutateAsync: updateMe } = useUpdateMeMutation();
  const form = useAppForm({
    defaultValues: {
      name: me.user.name,
      display_name: me.user.display_name,
      timezone: me.user.timezone || "UTC",
    },
    validators: { onChange: formSchema },
    onSubmit: async ({ value, formApi }) => {
      const parsed = formSchema.parse(value);
      await updateMe(parsed);
      formApi.reset(parsed);
    },
  });
  return (
    <SettingsSection
      title="Details"
      description={`Signed in as ${me.user.email}.`}
      onSubmit={submitHandler(form)}
      footer={
        <form.AppForm>
          <form.SubmitButton size="sm" requireDirty>
            Save
          </form.SubmitButton>
        </form.AppForm>
      }
    >
      <form.AppField
        name="name"
        children={(field) => <field.TextField label="Full name" autoComplete="name" />}
      />
      <form.AppField
        name="display_name"
        children={(field) => (
          <field.TextField
            label="Username"
            description="Others mention you with @ followed by this name."
            transform={(v) => v.toLowerCase()}
          />
        )}
      />
      <form.AppField
        name="timezone"
        children={(field) => <field.SelectField label="Timezone" items={timezones} />}
      />
    </SettingsSection>
  );
}
