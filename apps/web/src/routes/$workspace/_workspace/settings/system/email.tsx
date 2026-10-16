import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { SendIcon, Trash2Icon, ZapIcon } from "lucide-react";
import { z } from "zod";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  emailSettingsQuery,
  type TEmailEncryption,
  type TEmailSettings,
  useClearEmailSettingsMutation,
  useSendTestEmailMutation,
  useUpdateEmailSettingsMutation,
} from "@/lib/queries/email";

export const Route = createFileRoute("/$workspace/_workspace/settings/system/email")({
  component: SystemEmailSettingsPage,
});

const encryptionItems: { value: TEmailEncryption; label: string }[] = [
  { value: "starttls", label: "STARTTLS (port 587)" },
  { value: "tls", label: "TLS (port 465)" },
  { value: "none", label: "None (port 25)" },
];

const defaultPorts: Record<TEmailEncryption, string> = {
  starttls: "587",
  tls: "465",
  none: "25",
};

const resendPreset = {
  host: "smtp.resend.com",
  port: "465",
  encryption: "tls" as TEmailEncryption,
  username: "resend",
};

const formSchema = z.object({
  host: z.string().trim().min(1, "Host is required."),
  port: z
    .string()
    .refine((v) => Number.isInteger(Number(v)) && Number(v) >= 1 && Number(v) <= 65535, {
      message: "Port must be between 1 and 65535.",
    }),
  encryption: z.enum(["starttls", "tls", "none"]),
  username: z.string(),
  password: z.string(),
  from_address: z.string().trim().email("Enter a valid email address."),
  from_name: z.string(),
});

function SystemEmailSettingsPage() {
  const { data, isPending, error, refetch } = useQuery(emailSettingsQuery());
  return (
    <SettingsPage
      title="Email"
      description="Zigzag sends invitations, password resets and notifications through this SMTP relay."
    >
      {isPending && <ListSkeleton rows={4} />}
      {!isPending && error && <ErrorCard message={error.message} onRetry={() => void refetch()} />}
      {data && (
        <>
          <EmailSettingsForm
            key={JSON.stringify(data.data.settings ?? null)}
            settings={data.data}
          />
          {data.data.configured && <EmailActions />}
        </>
      )}
    </SettingsPage>
  );
}

function EmailSettingsForm({ settings }: { settings: TEmailSettings }) {
  const { mutateAsync: update } = useUpdateEmailSettingsMutation();
  const current = settings.settings;
  const form = useAppForm({
    defaultValues: {
      host: current?.host ?? "",
      port: String(current?.port ?? 587),
      encryption: (current?.encryption ?? "starttls") as TEmailEncryption,
      username: current?.username ?? "",
      password: "",
      from_address: current?.from_address ?? "",
      from_name: current?.from_name ?? "",
    },
    validators: { onChange: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      await update({
        host: parsed.host,
        port: Number(parsed.port),
        encryption: parsed.encryption,
        username: parsed.username,
        from_address: parsed.from_address,
        from_name: parsed.from_name,
        ...(parsed.password ? { password: parsed.password } : {}),
      });
    },
  });

  const applyResend = () => {
    form.setFieldValue("host", resendPreset.host);
    form.setFieldValue("port", resendPreset.port);
    form.setFieldValue("encryption", resendPreset.encryption);
    form.setFieldValue("username", resendPreset.username);
  };

  return (
    <SettingsSection
      title="SMTP relay"
      description="Set up SPF and DKIM for the sending domain at your provider, or messages may land in spam."
      actions={
        <>
          {settings.configured ? (
            <Badge variant="success">Configured</Badge>
          ) : (
            <Badge variant="muted">Not configured</Badge>
          )}
          <Button type="button" variant="outline" size="sm" onClick={applyResend}>
            <ZapIcon className="size-4" />
            Resend
          </Button>
        </>
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
        name="host"
        children={(field) => (
          <field.TextField label="Host" placeholder="smtp.example.com" autoComplete="off" />
        )}
      />
      <div className="grid gap-4 sm:grid-cols-2">
        <form.AppField
          name="encryption"
          listeners={{
            onChange: ({ value, fieldApi }) => {
              const previous = fieldApi.form.state.values.port;
              if (!Object.values(defaultPorts).includes(previous)) return;
              form.setFieldValue("port", defaultPorts[value]);
            },
          }}
          children={(field) => <field.SelectField label="Encryption" items={encryptionItems} />}
        />
        <form.AppField
          name="port"
          children={(field) => (
            <field.TextField label="Port" type="number" inputMode="numeric" min={1} max={65535} />
          )}
        />
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <form.AppField
          name="username"
          children={(field) => (
            <field.TextField
              label="Username"
              description="Leave blank for no authentication."
              autoComplete="off"
            />
          )}
        />
        <form.AppField
          name="password"
          children={(field) => (
            <field.TextField
              label="Password"
              type="password"
              autoComplete="new-password"
              description={
                settings.configured
                  ? "Leave blank to keep the stored password."
                  : "For Resend, use an API key."
              }
            />
          )}
        />
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <form.AppField
          name="from_address"
          children={(field) => (
            <field.TextField
              label="From address"
              type="email"
              placeholder="zigzag@example.com"
              autoComplete="off"
            />
          )}
        />
        <form.AppField
          name="from_name"
          children={(field) => (
            <field.TextField label="From name" placeholder="Zigzag" autoComplete="off" />
          )}
        />
      </div>
    </SettingsSection>
  );
}

function EmailActions() {
  const { mutate: sendTest, isPending } = useSendTestEmailMutation();
  const { mutateAsync: clear } = useClearEmailSettingsMutation();
  return (
    <SettingsSection
      title="Check and remove"
      description="Send yourself a test message, or remove the relay so Zigzag stops sending email."
      footer={
        <>
          <ConfirmDialog
            trigger={
              <Button variant="outline" size="sm">
                <Trash2Icon className="size-4" />
                Remove
              </Button>
            }
            title="Remove email settings"
            description="Zigzag stops sending email. Invitation links can still be copied and shared by hand."
            confirmLabel="Remove"
            onConfirm={() => clear()}
          />
          <Button
            size="sm"
            isPending={isPending}
            onClick={() =>
              sendTest(undefined, {
                onSuccess: () =>
                  toast.add({
                    type: "success",
                    title: "Test email sent",
                    description: "Check your inbox.",
                  }),
              })
            }
          >
            <SendIcon className="size-4" />
            Send test email
          </Button>
        </>
      }
    />
  );
}
