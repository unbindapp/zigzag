import { useQuery } from "@tanstack/react-query";
import { PlusIcon, XIcon } from "lucide-react";
import { ReactElement, useState } from "react";

import { CheckboxList } from "@/components/shared/checkbox-list";
import { MarkdownField } from "@/components/shared/markdown-editor";
import { PRIORITIES, PriorityIcon, priorityLabel } from "@/components/shared/priority-icon";
import { StateIcon } from "@/components/shared/state-icon";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useTeams } from "@/components/workspace/teams-provider";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { labelTreeQuery } from "@/lib/queries/labels";
import {
  type TTemplate,
  type TTemplateData,
  TemplateFormSchema,
  useCreateTemplateMutation,
  useUpdateTemplateMutation,
} from "@/lib/queries/templates";
import { workspaceMembersQuery } from "@/lib/queries/workspaces";

const NONE = "none";

type TProps = {
  trigger?: ReactElement;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  workspaceId: string;
  teamId?: string;
} & ({ mode: "create"; template?: never } | { mode: "edit"; template: TTemplate });

export function TemplateDialog(props: TProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = props.open ?? internalOpen;
  const setOpen = (next: boolean) => {
    setInternalOpen(next);
    props.onOpenChange?.(next);
  };
  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      {props.trigger && <DialogTrigger render={props.trigger} />}
      <DialogContent className="w-xl max-w-full">
        <DialogHeader>
          <DialogTitle>
            {props.mode === "edit" ? "Edit template" : "New issue template"}
          </DialogTitle>
        </DialogHeader>
        {isOpen && <TemplateForm {...props} onDone={() => setOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function initialValues(template?: TTemplate) {
  const data = template?.data;
  return {
    name: template?.name ?? "",
    description: template?.description ?? "",
    title: data?.title ?? "",
    body: data?.description ?? "",
    priority: data?.priority == null ? NONE : String(data.priority),
    stateId: data?.state_id ?? NONE,
    assigneeId: data?.assignee_id ?? NONE,
    labelIds: data?.label_ids ?? [],
    subIssues: (data?.sub_issues ?? []).map((sub) => sub.title),
  };
}

function optional(value: string) {
  return value === NONE ? undefined : value;
}

function TemplateForm(props: TProps & { onDone: () => void }) {
  const { mutateAsync: createTemplate } = useCreateTemplateMutation();
  const { mutateAsync: updateTemplate } = useUpdateTemplateMutation();
  const { teams } = useTeams();
  const teamId = props.mode === "edit" ? (props.template.team_id ?? undefined) : props.teamId;
  const team = teams.find((candidate) => candidate.id === teamId);
  const { data: labelData } = useQuery(labelTreeQuery({ workspaceId: props.workspaceId, teamId }));
  const { data: memberData } = useQuery(workspaceMembersQuery(props.workspaceId));

  const labelItems = (labelData?.labels ?? []).flatMap((label) =>
    (label.is_group ? label.children : [label]).map((item) => ({
      value: item.id,
      label: label.is_group ? `${label.name} / ${item.name}` : item.name,
      icon: (
        <span aria-hidden className="size-2 rounded-full" style={{ backgroundColor: item.color }} />
      ),
    })),
  );
  const priorityItems = [
    { value: NONE, label: "Not set" },
    ...PRIORITIES.map((priority) => ({
      value: String(priority),
      label: priorityLabel(priority),
      icon: <PriorityIcon priority={priority} />,
    })),
  ];
  const stateItems = [
    { value: NONE, label: "Team default" },
    ...(team?.states ?? []).map((state) => ({
      value: state.id,
      label: state.name,
      icon: <StateIcon type={state.type} color={state.color} />,
    })),
  ];
  const assigneeItems = [
    { value: NONE, label: "Unassigned" },
    ...(memberData?.members ?? []).map((member) => ({
      value: member.user.id,
      label: member.user.name,
    })),
  ];

  const form = useAppForm({
    defaultValues: initialValues(props.template),
    validators: { onSubmit: TemplateFormSchema },
    onSubmit: async ({ value }) => {
      const parsed = TemplateFormSchema.parse(value);
      const data: TTemplateData = {
        title: parsed.title || undefined,
        description: parsed.body || undefined,
        priority: parsed.priority === NONE ? undefined : Number(parsed.priority),
        state_id: optional(parsed.stateId),
        assignee_id: optional(parsed.assigneeId),
        label_ids: parsed.labelIds.length > 0 ? parsed.labelIds : undefined,
        sub_issues: parsed.subIssues
          .map((title) => title.trim())
          .filter(Boolean)
          .map((title) => ({ title })),
      };
      if (props.mode === "edit") {
        await updateTemplate({
          template_id: props.template.id,
          name: parsed.name,
          description: parsed.description,
          data,
        });
        props.onDone();
        return;
      }
      await createTemplate({
        workspace_id: props.workspaceId,
        team_id: teamId,
        kind: "issue",
        name: parsed.name,
        description: parsed.description || undefined,
        data,
      });
      props.onDone();
    },
  });

  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="name"
        children={(field) => <field.TextField label="Template name" autoFocus />}
      />
      <form.AppField
        name="description"
        children={(field) => <field.TextField label="When to use it" placeholder="Optional" />}
      />
      <form.AppField
        name="title"
        children={(field) => <field.TextField label="Issue title" placeholder="Bug: " />}
      />
      <form.Field
        name="body"
        children={(field) => (
          <MarkdownField
            label="Issue description"
            value={field.state.value}
            onChange={field.handleChange}
            placeholder="Steps to reproduce, expected result…"
          />
        )}
      />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <form.AppField
          name="priority"
          children={(field) => <field.SelectField label="Priority" items={priorityItems} />}
        />
        <form.AppField
          name="stateId"
          children={(field) => (
            <field.SelectField
              label="Status"
              items={stateItems}
              disabled={!team}
              description={team ? undefined : "Team templates only"}
            />
          )}
        />
        <form.AppField
          name="assigneeId"
          children={(field) => <field.SelectField label="Assignee" items={assigneeItems} />}
        />
      </div>
      <form.Field
        name="labelIds"
        children={(field) => (
          <div className="flex flex-col gap-1.5">
            <span className="text-sm font-medium">Labels</span>
            <CheckboxList
              label="Labels"
              items={labelItems}
              value={field.state.value}
              onChange={field.handleChange}
              emptyText="No labels yet."
            />
          </div>
        )}
      />
      <form.Field
        name="subIssues"
        children={(field) => (
          <div className="flex flex-col gap-2">
            <span className="text-sm font-medium">Sub-issues</span>
            {field.state.value.map((title, index) => (
              <div key={index} className="flex items-center gap-2">
                <Input
                  aria-label={`Sub-issue ${index + 1}`}
                  value={title}
                  placeholder="Sub-issue title"
                  onChange={(event) =>
                    field.handleChange(
                      field.state.value.map((current, i) =>
                        i === index ? event.target.value : current,
                      ),
                    )
                  }
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  aria-label={`Remove sub-issue ${index + 1}`}
                  onClick={() =>
                    field.handleChange(field.state.value.filter((_, i) => i !== index))
                  }
                >
                  <XIcon className="size-4" />
                </Button>
              </div>
            ))}
            <div className="flex">
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => field.handleChange([...field.state.value, ""])}
              >
                <PlusIcon className="size-4" />
                Add sub-issue
              </Button>
            </div>
          </div>
        )}
      />
      <DialogFooter>
        <DialogClose
          render={
            <Button type="button" variant="outline" size="sm">
              Cancel
            </Button>
          }
        />
        <form.AppForm>
          <form.SubmitButton size="sm">
            {props.mode === "edit" ? "Save" : "Create template"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
