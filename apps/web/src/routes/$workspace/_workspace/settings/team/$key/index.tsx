import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { z } from "zod";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { sanitizeTeamKey } from "@/lib/helpers/identifiers";
import { timezoneOptions } from "@/lib/helpers/timezones";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { TeamFormSchema, useUpdateTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/")({
  component: TeamGeneralSettingsPage,
});

const formSchema = TeamFormSchema.extend({
  description: z.string().max(255, "Keep the description under 255 characters."),
  icon: z.string(),
  color: z.string(),
  private: z.boolean(),
  timezone: z.string(),
});

const timezones = timezoneOptions();

function TeamGeneralSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { mutateAsync: updateTeam } = useUpdateTeamMutation();

  const form = useAppForm({
    defaultValues: {
      name: team.name,
      key: team.key,
      description: team.description ?? "",
      icon: team.icon ?? "",
      color: team.color ?? "",
      private: team.private,
      timezone: team.timezone || "UTC",
    },
    validators: { onChange: formSchema },
    onSubmit: async ({ value, formApi }) => {
      const parsed = formSchema.parse(value);
      const res = await updateTeam({
        team_id: team.id,
        name: parsed.name,
        key: parsed.key,
        description: parsed.description,
        icon: value.icon,
        color: value.color || undefined,
        private: value.private,
        timezone: value.timezone,
      });
      formApi.reset(value);
      if (res.data.key === team.key) return;
      await navigate({
        to: "/$workspace/settings/team/$key",
        params: { workspace: workspace.slug, key: res.data.key },
      });
    },
  });

  return (
    <SettingsPage title="General" description="Name, key and visibility of this team.">
      <SettingsSection
        title="Team"
        onSubmit={submitHandler(form)}
        footer={
          <form.AppForm>
            <form.SubmitButton size="sm" requireDirty>
              Save
            </form.SubmitButton>
          </form.AppForm>
        }
      >
        <div className="flex items-start gap-4">
          <form.Subscribe selector={(state) => state.values.color}>
            {(color) => (
              <form.AppField
                name="icon"
                children={(field) => <field.IconField label="Icon" color={color} />}
              />
            )}
          </form.Subscribe>
          <form.AppField name="color" children={(field) => <field.ColorField label="Color" />} />
        </div>
        <form.AppField name="name" children={(field) => <field.TextField label="Name" />} />
        <form.AppField
          name="key"
          children={(field) => (
            <field.TextField
              label="Key"
              description="Prefix of every issue identifier in this team."
              transform={sanitizeTeamKey}
              className="font-mono"
            />
          )}
        />
        <form.AppField
          name="description"
          children={(field) => <field.TextareaField label="Description" rows={3} />}
        />
        <form.AppField
          name="timezone"
          children={(field) => (
            <field.SelectField
              label="Timezone"
              description="Cycles start and end at midnight in this timezone."
              items={timezones}
            />
          )}
        />
        <form.AppField
          name="private"
          children={(field) => (
            <field.SwitchField
              label="Private team"
              description="Only members and workspace admins can see a private team."
            />
          )}
        />
      </SettingsSection>
    </SettingsPage>
  );
}
