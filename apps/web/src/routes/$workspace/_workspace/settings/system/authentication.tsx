import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Trash2Icon } from "lucide-react";
import { z } from "zod";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { CopyField } from "@/components/shared/copy-button";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  googleAuthSettingsQuery,
  type TGoogleAuthSettings,
  useClearGoogleAuthSettingsMutation,
  useUpdateGoogleAuthSettingsMutation,
} from "@/lib/queries/google-auth";

export const Route = createFileRoute("/$workspace/_workspace/settings/system/authentication")({
  component: SystemAuthenticationSettingsPage,
});

function formSchema(hasStoredSecret: boolean) {
  return z
    .object({
      enabled: z.boolean(),
      client_id: z.string().trim().min(1, "Client ID is required."),
      client_secret: z.string().trim(),
      hosted_domain: z
        .string()
        .trim()
        .toLowerCase()
        .refine((v) => v === "" || /^[a-z0-9-]+(\.[a-z0-9-]+)+$/.test(v), {
          message: "Enter a domain such as example.com.",
        }),
    })
    .refine((v) => !v.enabled || hasStoredSecret || v.client_secret !== "", {
      path: ["client_secret"],
      message: "Add the client secret to turn Google sign-in on.",
    });
}

function SystemAuthenticationSettingsPage() {
  const { data, isPending, error, refetch } = useQuery(googleAuthSettingsQuery());
  return (
    <SettingsPage
      title="Authentication"
      description="Let people sign in with their Google account next to email and password."
    >
      {isPending && <ListSkeleton rows={4} />}
      {!isPending && error && <ErrorCard message={error.message} onRetry={() => void refetch()} />}
      {data && (
        <>
          <GoogleAuthForm key={JSON.stringify(data.data)} settings={data.data} />
          <GoogleCloudSteps redirectUri={data.data.redirect_uri} />
          {data.data.configured && <RemoveGoogleAuth />}
        </>
      )}
    </SettingsPage>
  );
}

function GoogleAuthForm({ settings }: { settings: TGoogleAuthSettings }) {
  const { mutateAsync: update } = useUpdateGoogleAuthSettingsMutation();
  const schema = formSchema(settings.configured);
  const form = useAppForm({
    defaultValues: {
      enabled: settings.enabled,
      client_id: settings.client_id,
      client_secret: "",
      hosted_domain: settings.hosted_domain,
    },
    validators: { onChange: schema },
    onSubmit: async ({ value }) => {
      const parsed = schema.parse(value);
      await update({
        enabled: parsed.enabled,
        client_id: parsed.client_id,
        hosted_domain: parsed.hosted_domain,
        ...(parsed.client_secret ? { client_secret: parsed.client_secret } : {}),
      });
    },
  });

  return (
    <SettingsSection
      title="Sign in with Google"
      description="Accounts are matched by verified email. New people get an account only when they have an invitation or public signup is on."
      actions={
        settings.enabled ? <Badge variant="success">On</Badge> : <Badge variant="muted">Off</Badge>
      }
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
        name="enabled"
        children={(field) => (
          <field.SwitchField
            label="Enable Google sign-in"
            description="Shows Continue with Google on the sign-in page."
          />
        )}
      />
      <form.AppField
        name="client_id"
        children={(field) => (
          <field.TextField
            label="Client ID"
            placeholder="1234567890-abc.apps.googleusercontent.com"
            autoComplete="off"
          />
        )}
      />
      <form.AppField
        name="client_secret"
        children={(field) => (
          <field.TextField
            label="Client secret"
            type="password"
            autoComplete="new-password"
            placeholder={settings.secret_prefix ? `${settings.secret_prefix}…` : "GOCSPX-…"}
            description={
              settings.secret_prefix
                ? `Stored, starts with ${settings.secret_prefix}. Leave blank to keep it.`
                : "Only its first characters are shown after saving."
            }
          />
        )}
      />
      <form.AppField
        name="hosted_domain"
        children={(field) => (
          <field.TextField
            label="Hosted domain"
            placeholder="example.com"
            autoComplete="off"
            description="Optional. Only Google Workspace accounts of this domain can sign in; leave blank to allow any Google account."
          />
        )}
      />
    </SettingsSection>
  );
}

function GoogleCloudSteps({ redirectUri }: { redirectUri: string }) {
  return (
    <SettingsSection
      title="Set up in Google Cloud"
      description="Create an OAuth client for this instance in the Google Cloud console."
    >
      <div className="flex flex-col gap-1.5">
        <p className="text-sm font-medium">Redirect URI</p>
        <CopyField value={redirectUri} label="Redirect URI" />
      </div>
      <ol className="text-muted-foreground flex list-decimal flex-col gap-1 pl-4 text-sm">
        <li>Open the Google Cloud console and pick or create a project.</li>
        <li>Configure the OAuth consent screen with your app name and support email.</li>
        <li>Create an OAuth client ID of type Web application.</li>
        <li>Add the redirect URI above to its authorized redirect URIs.</li>
        <li>Paste the client ID and client secret here, turn Google sign-in on and save.</li>
      </ol>
    </SettingsSection>
  );
}

function RemoveGoogleAuth() {
  const { mutateAsync: clear } = useClearGoogleAuthSettingsMutation();
  return (
    <SettingsSection
      title="Remove"
      description="Delete the client ID and secret. Google sign-in stops; accounts keep their passwords and their Google link."
      footer={
        <ConfirmDialog
          trigger={
            <Button variant="outline" size="sm">
              <Trash2Icon className="size-4" />
              Remove
            </Button>
          }
          title="Remove Google sign-in"
          description="People who only sign in with Google need a password reset to get back in."
          confirmLabel="Remove Google sign-in"
          onConfirm={() => clear()}
        />
      }
    />
  );
}
