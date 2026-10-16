import { useQuery } from "@tanstack/react-query";
import { EllipsisIcon, FileStackIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import { List, ListRow, ListRowText } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { SettingsSection } from "@/components/layout/settings-section";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { PriorityIcon } from "@/components/shared/priority-icon";
import { TemplateDialog } from "@/components/templates/template-dialog";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  type TTemplate,
  templatesListQuery,
  useDeleteTemplateMutation,
} from "@/lib/queries/templates";

type TProps = {
  workspaceId: string;
  teamId?: string;
  canEdit: boolean;
  title: string;
  description: string;
};

export function TemplateManager({ workspaceId, teamId, canEdit, title, description }: TProps) {
  const { data, isPending, error, refetch } = useQuery(templatesListQuery(workspaceId, teamId));
  const templates = (data?.templates ?? []).filter(
    (template) => (template.team_id ?? null) === (teamId ?? null),
  );
  const createButton = canEdit && (
    <TemplateDialog
      mode="create"
      workspaceId={workspaceId}
      teamId={teamId}
      trigger={
        <Button size="sm">
          <PlusIcon className="size-4" />
          New template
        </Button>
      }
    />
  );
  return (
    <SettingsSection title={title} description={description} actions={createButton}>
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={() => void refetch()}
        isEmpty={templates.length === 0}
        empty={{
          icon: FileStackIcon,
          title: "No templates yet",
          description:
            "Templates prefill the title, description, properties and sub-issues of new issues.",
        }}
      >
        <List>
          {templates.map((template) => (
            <TemplateRow key={template.id} template={template} canEdit={canEdit} />
          ))}
        </List>
      </ListStates>
    </SettingsSection>
  );
}

function TemplateRow({ template, canEdit }: { template: TTemplate; canEdit: boolean }) {
  const [dialog, setDialog] = useState<"edit" | "delete" | null>(null);
  const { mutateAsync: deleteTemplate } = useDeleteTemplateMutation();
  const subIssues = template.data.sub_issues?.length ?? 0;
  const close = (open: boolean) => {
    if (!open) setDialog(null);
  };
  return (
    <ListRow
      actions={
        canEdit && (
          <>
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button variant="ghost" size="icon" aria-label={`Actions for ${template.name}`}>
                    <EllipsisIcon className="size-4" />
                  </Button>
                }
              />
              <DropdownMenuContent align="end" className="w-44">
                <DropdownMenuItem onClick={() => setDialog("edit")}>
                  <PencilIcon className="size-4" />
                  Edit template
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="text-destructive" onClick={() => setDialog("delete")}>
                  <Trash2Icon className="size-4" />
                  Delete template
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <TemplateDialog
              mode="edit"
              template={template}
              workspaceId={template.workspace_id}
              open={dialog === "edit"}
              onOpenChange={close}
            />
            <ConfirmDialog
              open={dialog === "delete"}
              onOpenChange={close}
              title={`Delete ${template.name}`}
              description="Issues created from this template stay as they are."
              confirmLabel="Delete template"
              onConfirm={() => deleteTemplate({ templateId: template.id })}
            />
          </>
        )
      }
    >
      <FileStackIcon className="text-muted-foreground size-4" />
      <ListRowText
        primary={template.name}
        secondary={template.description || template.data.title}
      />
      {template.data.priority != null && template.data.priority > 0 && (
        <PriorityIcon priority={template.data.priority} />
      )}
      {subIssues > 0 && (
        <span className="text-muted-foreground text-xs">
          {subIssues} {subIssues === 1 ? "sub-issue" : "sub-issues"}
        </span>
      )}
    </ListRow>
  );
}
