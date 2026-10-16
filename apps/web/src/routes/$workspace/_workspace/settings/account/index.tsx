import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { MonitorIcon, MoonIcon, SunIcon, UploadIcon, XIcon } from "lucide-react";
import { useTheme } from "next-themes";
import { type ReactNode } from "react";
import { z } from "zod";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { PropertyRow } from "@/components/layout/property-row";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { defaultTheme, type TTheme } from "@/components/providers/themes";
import { githubLoginSchema } from "@/components/integrations/github-sync";
import { Avatar } from "@/components/shared/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { FileTrigger } from "@/components/ui/file-trigger";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { timezoneOptions } from "@/lib/helpers/timezones";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { meQuery, NO_AVATAR_UPLOAD_ID, type TMe, useUpdateMeMutation } from "@/lib/queries/me";
import { imageUploadAccept, useCreateUploadMutation } from "@/lib/queries/uploads";

export const Route = createFileRoute("/$workspace/_workspace/settings/account/")({
  component: AccountProfileSettingsPage,
});

const timezones = timezoneOptions();

const themeOptions: { value: TTheme; label: string; icon: ReactNode }[] = [
  { value: "system", label: "System", icon: <MonitorIcon className="size-4" /> },
  { value: "light", label: "Light", icon: <SunIcon className="size-4" /> },
  { value: "dark", label: "Dark", icon: <MoonIcon className="size-4" /> },
];

const formSchema = z.object({
  name: z.string().trim().min(1, "Name is required."),
  display_name: z
    .string()
    .trim()
    .regex(/^[a-z0-9._-]{2,32}$/, "Use 2 to 32 lowercase letters, numbers, dots or dashes."),
  timezone: z.string().min(1),
  github_login: githubLoginSchema,
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
          <SignInMethodsSection me={me} />
          <AppearanceSection />
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
          accept={imageUploadAccept}
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
      github_login: me.user.github_login ?? "",
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
      <form.AppField
        name="github_login"
        children={(field) => (
          <field.TextField
            label="GitHub username"
            placeholder="octocat"
            autoComplete="off"
            description="Issues and comments you write on synced repositories show as yours, and assigning you here assigns you on GitHub."
          />
        )}
      />
    </SettingsSection>
  );
}

function SignInMethodsSection({ me }: { me: TMe }) {
  return (
    <SettingsSection title="Sign-in methods" description="Ways you can sign in to this account.">
      <PropertyRow label="Password">
        {me.user.has_password ? (
          <Badge variant="success">Set</Badge>
        ) : (
          <>
            <Badge variant="muted">Not set</Badge>
            <Link
              to="/forgot-password"
              className="text-muted-foreground has-hover:hover:text-foreground text-sm underline-offset-4 has-hover:hover:underline"
            >
              Set one with a reset link
            </Link>
          </>
        )}
      </PropertyRow>
      <PropertyRow label="Google">
        {me.user.google_linked ? (
          <Badge variant="success">Linked</Badge>
        ) : (
          <Badge variant="muted">Not linked</Badge>
        )}
      </PropertyRow>
    </SettingsSection>
  );
}

function AppearanceSection() {
  const { theme, setTheme } = useTheme();
  return (
    <SettingsSection title="Appearance" description="Saved in this browser.">
      <ThemeField key={theme} theme={theme ?? defaultTheme} onChange={setTheme} />
    </SettingsSection>
  );
}

function ThemeField({ theme, onChange }: { theme: string; onChange: (theme: string) => void }) {
  const form = useAppForm({ defaultValues: { theme } });
  return (
    <form.AppField
      name="theme"
      listeners={{ onChange: ({ value }) => onChange(value) }}
      children={(field) => <field.SelectField label="Theme" items={themeOptions} />}
    />
  );
}
