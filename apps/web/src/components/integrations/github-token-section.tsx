import { KeyRoundIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";
import { z } from "zod";

import { githubTokenSchema } from "@/components/integrations/github-sync";
import { PropertyRow } from "@/components/layout/property-row";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { type TGithubIntegration, useUpdateGithubTokenMutation } from "@/lib/queries/integrations";

const title = "Issues sync token";
const description =
  "A fine-grained token with Issues read and write on the synced repositories. Zigzag uses it to send changes back to GitHub and to import existing issues.";

const formSchema = z.object({ token: githubTokenSchema });

export function GithubTokenSection({ integration }: { integration: TGithubIntegration }) {
  const [isReplacing, setIsReplacing] = useState(false);
  if (!integration.has_token || isReplacing) {
    return (
      <GithubTokenForm hasToken={integration.has_token} onDone={() => setIsReplacing(false)} />
    );
  }
  return <StoredGithubToken integration={integration} onReplace={() => setIsReplacing(true)} />;
}

function StoredGithubToken({
  integration,
  onReplace,
}: {
  integration: TGithubIntegration;
  onReplace: () => void;
}) {
  const { workspace } = useWorkspace();
  const { mutateAsync: update } = useUpdateGithubTokenMutation();
  return (
    <SettingsSection
      title={title}
      description={description}
      actions={
        <>
          <Button variant="outline" size="sm" onClick={onReplace}>
            <KeyRoundIcon className="size-4" />
            Replace
          </Button>
          <ConfirmDialog
            trigger={
              <Button variant="outline" size="sm">
                <Trash2Icon className="size-4" />
                Clear
              </Button>
            }
            title="Clear the GitHub token"
            description="Changes made in Zigzag stop going to GitHub and imports stop working. Changes from GitHub keep coming in through the webhook."
            confirmLabel="Clear token"
            onConfirm={() => update({ workspaceId: workspace.id, token: "" })}
          />
        </>
      }
    >
      <PropertyRow label="Token">
        <span className="font-mono text-xs">{integration.token_prefix}…</span>
        <Badge variant="success">Stored</Badge>
      </PropertyRow>
    </SettingsSection>
  );
}

function GithubTokenForm({ hasToken, onDone }: { hasToken: boolean; onDone: () => void }) {
  const { workspace } = useWorkspace();
  const { mutateAsync: update } = useUpdateGithubTokenMutation();
  const form = useAppForm({
    defaultValues: { token: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      await update({ workspaceId: workspace.id, token: formSchema.parse(value).token });
      onDone();
    },
  });
  return (
    <SettingsSection
      title={title}
      description={description}
      onSubmit={submitHandler(form)}
      footer={
        <>
          {hasToken && (
            <Button type="button" variant="outline" size="sm" onClick={onDone}>
              Cancel
            </Button>
          )}
          <form.AppForm>
            <form.SubmitButton size="sm">Save token</form.SubmitButton>
          </form.AppForm>
        </>
      }
    >
      <form.AppField
        name="token"
        children={(field) => (
          <field.TextField
            label="Token"
            type="password"
            autoComplete="off"
            placeholder="github_pat_…"
            description="On GitHub: Settings, Developer settings, Fine-grained tokens. Only its first characters are shown after saving."
          />
        )}
      />
    </SettingsSection>
  );
}
