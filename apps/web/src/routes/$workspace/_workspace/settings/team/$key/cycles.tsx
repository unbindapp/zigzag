import { createFileRoute } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useUpdateTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/cycles")({
  component: TeamCyclesSettingsPage,
});

const weeks = (from: number, to: number) =>
  Array.from({ length: to - from + 1 }, (_, i) => from + i).map((n) => ({
    value: String(n),
    label: n === 0 ? "No cooldown" : `${n} ${n === 1 ? "week" : "weeks"}`,
  }));

const durationItems = weeks(1, 8);
const cooldownItems = weeks(0, 4);
const dayItems = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"].map(
  (label, index) => ({ value: String(index + 1), label }),
);

function TeamCyclesSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { mutateAsync: updateTeam } = useUpdateTeamMutation();

  const form = useAppForm({
    defaultValues: {
      cycles_enabled: team.cycles_enabled,
      cycle_duration_weeks: String(team.cycle_duration_weeks),
      cycle_cooldown_weeks: String(team.cycle_cooldown_weeks),
      cycle_start_day: String(team.cycle_start_day),
      cycle_auto_move: team.cycle_auto_move,
      cycle_auto_add_started: team.cycle_auto_add_started,
    },
    onSubmit: async ({ value, formApi }) => {
      await updateTeam({
        team_id: team.id,
        cycles_enabled: value.cycles_enabled,
        cycle_duration_weeks: Number(value.cycle_duration_weeks),
        cycle_cooldown_weeks: Number(value.cycle_cooldown_weeks),
        cycle_start_day: Number(value.cycle_start_day),
        cycle_auto_move: value.cycle_auto_move,
        cycle_auto_add_started: value.cycle_auto_add_started,
      });
      formApi.reset(value);
    },
  });

  return (
    <SettingsPage title="Cycles" description="Time boxed iterations for this team.">
      <SettingsSection
        title="Cycles"
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
          name="cycles_enabled"
          children={(field) => (
            <field.SwitchField
              label="Enable cycles"
              description="New cycles are created automatically on the schedule below."
            />
          )}
        />
        <form.Subscribe selector={(state) => state.values.cycles_enabled}>
          {(enabled) => (
            <>
              <form.AppField
                name="cycle_duration_weeks"
                children={(field) => (
                  <field.SelectField label="Duration" items={durationItems} disabled={!enabled} />
                )}
              />
              <form.AppField
                name="cycle_cooldown_weeks"
                children={(field) => (
                  <field.SelectField
                    label="Cooldown"
                    description="A break between cycles."
                    items={cooldownItems}
                    disabled={!enabled}
                  />
                )}
              />
              <form.AppField
                name="cycle_start_day"
                children={(field) => (
                  <field.SelectField label="Start day" items={dayItems} disabled={!enabled} />
                )}
              />
              <form.AppField
                name="cycle_auto_move"
                children={(field) => (
                  <field.SwitchField
                    label="Move unfinished issues"
                    description="Issues that are not done move to the next cycle when a cycle ends."
                    disabled={!enabled}
                  />
                )}
              />
              <form.AppField
                name="cycle_auto_add_started"
                children={(field) => (
                  <field.SwitchField
                    label="Add started issues"
                    description="An issue that moves to a started state joins the active cycle."
                    disabled={!enabled}
                  />
                )}
              />
            </>
          )}
        </form.Subscribe>
      </SettingsSection>
    </SettingsPage>
  );
}
