import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  notificationSettingsQuery,
  type TNotificationSettings,
  useUpdateNotificationSettingsMutation,
} from "@/lib/queries/me";

export const Route = createFileRoute("/$workspace/_workspace/settings/account/notifications")({
  component: AccountNotificationsSettingsPage,
});

const typeFields = [
  { name: "assigned", label: "Assigned to me", description: "An issue is assigned to you." },
  { name: "mentioned", label: "Mentions", description: "Someone mentions you." },
  { name: "commented", label: "Comments", description: "Someone comments on an issue you follow." },
  {
    name: "state_changed",
    label: "Status changes",
    description: "An issue you follow changes state.",
  },
  { name: "due_soon", label: "Due soon", description: "An issue assigned to you is due soon." },
  {
    name: "project_updates",
    label: "Project updates",
    description: "A project you are in posts an update.",
  },
] as const;

function AccountNotificationsSettingsPage() {
  const { data, isPending, error, refetch } = useQuery(notificationSettingsQuery());
  return (
    <SettingsPage title="Notifications" description="Choose what reaches your inbox and email.">
      {isPending && <ListSkeleton rows={4} />}
      {!isPending && error && <ErrorCard message={error.message} onRetry={() => void refetch()} />}
      {data && <NotificationsForm settings={data.settings} />}
    </SettingsPage>
  );
}

function NotificationsForm({ settings }: { settings: TNotificationSettings }) {
  const { mutateAsync: update } = useUpdateNotificationSettingsMutation();
  const form = useAppForm({
    defaultValues: settings,
    onSubmit: async ({ value, formApi }) => {
      const res = await update(value);
      formApi.reset(res.data);
    },
  });
  return (
    <SettingsSection
      title="Notify me about"
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
        name="email_enabled"
        children={(field) => (
          <field.SwitchField
            label="Email"
            description="Also send notifications by email, batched every few minutes."
          />
        )}
      />
      {typeFields.map((item) => (
        <form.AppField
          key={item.name}
          name={item.name}
          children={(field) => (
            <field.SwitchField label={item.label} description={item.description} />
          )}
        />
      ))}
    </SettingsSection>
  );
}
