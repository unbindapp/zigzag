import { createFileRoute } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useUpdateTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/auto-archive")({
  component: TeamAutoArchiveSettingsPage,
});

const NEVER = "0";
const DEFAULT_ARCHIVE_MONTHS = 6;
const months = [1, 3, 6, 9, 12, 24].map((n) => ({
  value: String(n),
  label: `After ${n} ${n === 1 ? "month" : "months"}`,
}));
const closeItems = [{ value: NEVER, label: "Never" }, ...months];
const archiveItems = months;

function TeamAutoArchiveSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { mutateAsync: updateTeam } = useUpdateTeamMutation();
  const form = useAppForm({
    defaultValues: {
      auto_close_months: String(team.auto_close_months ?? 0),
      auto_archive_months: String(team.auto_archive_months ?? DEFAULT_ARCHIVE_MONTHS),
    },
    onSubmit: async ({ value, formApi }) => {
      await updateTeam({
        team_id: team.id,
        auto_close_months: Number(value.auto_close_months),
        auto_archive_months: Number(value.auto_archive_months),
      });
      formApi.reset(value);
    },
  });

  return (
    <SettingsPage title="Auto close & archive" description="Keep the backlog and history tidy.">
      <SettingsSection
        title="Automation"
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
          name="auto_close_months"
          children={(field) => (
            <field.SelectField
              label="Auto close"
              description="Close issues that have not been updated for this long."
              items={closeItems}
            />
          )}
        />
        <form.AppField
          name="auto_archive_months"
          children={(field) => (
            <field.SelectField
              label="Auto archive"
              description="Archive completed and canceled issues after this long."
              items={archiveItems}
            />
          )}
        />
      </SettingsSection>
    </SettingsPage>
  );
}
