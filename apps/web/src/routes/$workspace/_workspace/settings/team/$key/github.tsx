import { createFileRoute, Link } from "@tanstack/react-router";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { StateIcon } from "@/components/shared/state-icon";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm, type TSelectFieldItem } from "@/lib/hooks/use-app-form";
import { CLEAR_UUID } from "@/lib/queries/issues";
import { type TTeam, useUpdateTeamMutation } from "@/lib/queries/teams";

export const Route = createFileRoute("/$workspace/_workspace/settings/team/$key/github")({
  component: TeamGithubSettingsPage,
});

const DEFAULT_STATE = "default";

function stateItems(team: TTeam, fallbackType: "started" | "completed"): TSelectFieldItem[] {
  const fallback = team.states.find((state) => state.type === fallbackType);
  return [
    {
      value: DEFAULT_STATE,
      label: fallback ? `Default (${fallback.name})` : "Default (no change)",
    },
    ...team.states.map((state) => ({
      value: state.id,
      label: state.name,
      icon: <StateIcon type={state.type} color={state.color} name={state.name} />,
    })),
  ];
}

function TeamGithubSettingsPage() {
  const team = useRequiredTeamFromRoute();
  const { workspace } = useWorkspace();
  const { mutateAsync: updateTeam } = useUpdateTeamMutation();
  const toInput = (value: string) => (value === DEFAULT_STATE ? CLEAR_UUID : value);
  const form = useAppForm({
    defaultValues: {
      opened: team.github_pr_opened_state_id ?? DEFAULT_STATE,
      merged: team.github_pr_merged_state_id ?? DEFAULT_STATE,
    },
    onSubmit: async ({ value, formApi }) => {
      await updateTeam({
        team_id: team.id,
        github_pr_opened_state_id: toInput(value.opened),
        github_pr_merged_state_id: toInput(value.merged),
      });
      formApi.reset(value);
    },
  });

  return (
    <SettingsPage title="GitHub" description="Move issues when their pull requests change.">
      <SettingsSection
        title="Pull request automation"
        description="By default an opened pull request moves its issues to the team's first started state, and a merged one moves the issues it fixes to the first completed state."
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
          name="opened"
          children={(field) => (
            <field.SelectField
              label="Pull request opened"
              description="For triage, backlog and unstarted issues, when a pull request is opened or ready for review or a branch named after the issue is pushed."
              items={stateItems(team, "started")}
            />
          )}
        />
        <form.AppField
          name="merged"
          children={(field) => (
            <field.SelectField
              label="Pull request merged"
              description={`For issues the pull request fixes, closes or resolves, like Fixes ${team.key}-123, unless they are completed or canceled.`}
              items={stateItems(team, "completed")}
            />
          )}
        />
        <p className="text-muted-foreground text-sm">
          Connect GitHub in{" "}
          <Link
            to="/$workspace/settings/integrations"
            params={{ workspace: workspace.slug }}
            className="text-foreground underline-offset-4 has-hover:hover:underline"
          >
            workspace integrations
          </Link>
          .
        </p>
      </SettingsSection>
    </SettingsPage>
  );
}
