import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  systemSettingsQuery,
  type TSystemSettings,
  useUpdateSystemSettingsMutation,
} from "@/lib/queries/system";

export const Route = createFileRoute("/$workspace/_workspace/settings/system/")({
  component: SystemGeneralSettingsPage,
});

const formSchema = z.object({
  instance_name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(64, "Keep it under 64 characters."),
  allow_public_signup: z.boolean(),
});

function SystemGeneralSettingsPage() {
  const { data, isPending, error, refetch } = useQuery(systemSettingsQuery());
  return (
    <SettingsPage title="General" description="Settings for this Zigzag instance.">
      {isPending && <ListSkeleton rows={2} />}
      {!isPending && error && <ErrorCard message={error.message} onRetry={() => void refetch()} />}
      {data && <SystemGeneralForm settings={data.settings} />}
    </SettingsPage>
  );
}

function SystemGeneralForm({ settings }: { settings: TSystemSettings }) {
  const { mutateAsync: update } = useUpdateSystemSettingsMutation();
  const form = useAppForm({
    defaultValues: {
      instance_name: settings.instance_name,
      allow_public_signup: settings.allow_public_signup,
    },
    validators: { onChange: formSchema },
    onSubmit: async ({ value, formApi }) => {
      const parsed = formSchema.parse(value);
      await update(parsed);
      formApi.reset(parsed);
    },
  });
  return (
    <SettingsSection
      title="Instance"
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
        name="instance_name"
        children={(field) => (
          <field.TextField
            label="Instance name"
            description="Shown in emails and the browser title."
          />
        )}
      />
      <form.AppField
        name="allow_public_signup"
        children={(field) => (
          <field.SwitchField
            label="Allow public signup"
            description="Anyone who can reach this instance can create an account."
          />
        )}
      />
    </SettingsSection>
  );
}
