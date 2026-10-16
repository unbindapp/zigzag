import { createFileRoute } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { EstimateBadge } from "@/components/shared/estimate-badge";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { estimateOptions, estimateTypeLabels, type TEstimateType } from "@/lib/helpers/estimates";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useUpdateTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/estimates")({
  component: TeamEstimatesSettingsPage,
});

const typeItems = (Object.keys(estimateTypeLabels) as TEstimateType[]).map((type) => ({
  value: type,
  label: estimateTypeLabels[type],
}));

function TeamEstimatesSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { mutateAsync: updateTeam } = useUpdateTeamMutation();

  const form = useAppForm({
    defaultValues: {
      estimate_type: team.estimate_type as TEstimateType,
      estimate_allow_zero: team.estimate_allow_zero,
      estimate_extended: team.estimate_extended,
    },
    onSubmit: async ({ value, formApi }) => {
      await updateTeam({ team_id: team.id, ...value });
      formApi.reset(value);
    },
  });

  return (
    <SettingsPage title="Estimates" description="How this team sizes its issues.">
      <SettingsSection
        title="Estimates"
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
          name="estimate_type"
          children={(field) => <field.SelectField label="Scale" items={typeItems} />}
        />
        <form.Subscribe selector={(state) => state.values}>
          {(values) => (
            <>
              <form.AppField
                name="estimate_allow_zero"
                children={(field) => (
                  <field.SwitchField
                    label="Allow zero"
                    description="Lets an issue be estimated at zero."
                    disabled={values.estimate_type === "none"}
                  />
                )}
              />
              <form.AppField
                name="estimate_extended"
                children={(field) => (
                  <field.SwitchField
                    label="Extended scale"
                    description="Adds two larger values to the scale."
                    disabled={values.estimate_type === "none"}
                  />
                )}
              />
              <EstimatePreview
                type={values.estimate_type}
                allowZero={values.estimate_allow_zero}
                extended={values.estimate_extended}
              />
            </>
          )}
        </form.Subscribe>
      </SettingsSection>
    </SettingsPage>
  );
}

function EstimatePreview({
  type,
  allowZero,
  extended,
}: {
  type: TEstimateType;
  allowZero: boolean;
  extended: boolean;
}) {
  const options = estimateOptions({ type, allowZero, extended });
  if (options.length === 0) {
    return <p className="text-muted-foreground text-sm">Estimates are turned off for this team.</p>;
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
      {options.map((value) => (
        <EstimateBadge key={value} value={value} type={type} />
      ))}
    </div>
  );
}
