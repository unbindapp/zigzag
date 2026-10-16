import { createFileRoute } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useUpdateTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/triage")({
  component: TeamTriageSettingsPage,
});

function TeamTriageSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { mutateAsync: updateTeam } = useUpdateTeamMutation();
  const form = useAppForm({
    defaultValues: { triage_enabled: team.triage_enabled },
    onSubmit: async ({ value, formApi }) => {
      await updateTeam({ team_id: team.id, triage_enabled: value.triage_enabled });
      formApi.reset(value);
    },
  });

  return (
    <SettingsPage title="Triage" description="Review incoming issues before they reach the team.">
      <SettingsSection
        title="Triage"
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
          name="triage_enabled"
          children={(field) => (
            <field.SwitchField
              label="Enable triage"
              description="Issues created by people outside the team, by integrations and by imports land in Triage first."
            />
          )}
        />
      </SettingsSection>
    </SettingsPage>
  );
}
