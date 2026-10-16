import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { z } from "zod";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useUpdateWorkspaceMutation, WorkspaceFormSchema } from "@/lib/queries/workspaces";

export const Route = createFileRoute("/$workspace/_workspace/settings/_admin/")({
  component: WorkspaceGeneralSettingsPage,
});

const formSchema = WorkspaceFormSchema.extend({ icon: z.string(), color: z.string() });

function WorkspaceGeneralSettingsPage() {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { mutateAsync: update } = useUpdateWorkspaceMutation();

  const form = useAppForm({
    defaultValues: {
      name: workspace.name,
      slug: workspace.slug,
      icon: workspace.icon ?? "",
      color: workspace.color ?? "",
    },
    validators: { onChange: formSchema },
    onSubmit: async ({ value, formApi }) => {
      const parsed = formSchema.parse(value);
      const res = await update({
        workspace_id: workspace.id,
        name: parsed.name,
        slug: parsed.slug,
        icon: value.icon,
        color: value.color || undefined,
      });
      formApi.reset({
        name: res.data.name,
        slug: res.data.slug,
        icon: res.data.icon ?? "",
        color: res.data.color ?? "",
      });
      if (res.data.slug === workspace.slug) return;
      await navigate({ to: "/$workspace/settings", params: { workspace: res.data.slug } });
    },
  });

  return (
    <SettingsPage title="General" description="Name, URL and look of this workspace.">
      <SettingsSection
        title="Workspace"
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
          name="slug"
          children={(field) => (
            <field.TextField
              label="URL"
              description={`${window.location.origin}/${field.state.value || "…"}`}
              transform={(v) => v.toLowerCase()}
            />
          )}
        />
      </SettingsSection>
    </SettingsPage>
  );
}
