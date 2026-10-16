import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { UploadIcon, XIcon } from "lucide-react";
import { z } from "zod";

import { SettingsPage } from "@/components/layout/settings-page";
import { SettingsSection } from "@/components/layout/settings-section";
import { Button } from "@/components/ui/button";
import { FileTrigger } from "@/components/ui/file-trigger";
import { WorkspaceIcon } from "@/components/workspace/workspace-icon";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { useCreateUploadMutation } from "@/lib/queries/uploads";
import {
  NO_WORKSPACE_ICON_UPLOAD_ID,
  useUpdateWorkspaceMutation,
  WorkspaceFormSchema,
} from "@/lib/queries/workspaces";

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
          <WorkspaceImageField />
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

function WorkspaceImageField() {
  const { workspace } = useWorkspace();
  const { mutateAsync: upload, isPending: isUploading } = useCreateUploadMutation();
  const { mutateAsync: update, isPending: isSaving } = useUpdateWorkspaceMutation();
  const setImage = (uploadId: string) =>
    update({ workspace_id: workspace.id, icon_upload_id: uploadId }).catch(() => undefined);
  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <p className="text-sm font-medium">Image</p>
      <div className="flex items-center gap-2">
        {workspace.icon_url && (
          <WorkspaceIcon
            id={workspace.id}
            name={workspace.name}
            icon={workspace.icon}
            color={workspace.color}
            iconUrl={workspace.icon_url}
          />
        )}
        <FileTrigger
          accept="image/*"
          onSelect={async ([file]) => {
            const res = await upload({ workspaceId: workspace.id, file }).catch(() => null);
            if (!res) return;
            await setImage(res.data.id);
          }}
        >
          {(open) => (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={open}
              isPending={isUploading || isSaving}
            >
              <UploadIcon className="size-4" />
              Upload image
            </Button>
          )}
        </FileTrigger>
        {workspace.icon_url && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => void setImage(NO_WORKSPACE_ICON_UPLOAD_ID)}
          >
            <XIcon className="size-4" />
            Remove image
          </Button>
        )}
      </div>
    </div>
  );
}
