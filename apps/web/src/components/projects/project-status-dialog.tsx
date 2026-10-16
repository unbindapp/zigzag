import { type ReactElement, useState } from "react";
import { z } from "zod";

import { projectStatusTypeLabels } from "@/components/projects/project-status-icon";
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
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  type TProjectStatus,
  type TProjectStatusType,
  useCreateProjectStatusMutation,
  useUpdateProjectStatusMutation,
} from "@/lib/queries/projects";

export const defaultProjectStatusColors: Record<TProjectStatusType, string> = {
  backlog: "#95a2b3",
  planned: "#e2e2e2",
  started: "#f2c94c",
  paused: "#f2994a",
  completed: "#5e6ad2",
  canceled: "#95a2b3",
};

const formSchema = z.object({
  name: z.string().trim().min(1, "Name is required.").max(64, "Keep the name under 64 characters."),
  color: z.string(),
  description: z.string().max(500, "Keep the description under 500 characters."),
});

type TProps = { trigger: ReactElement } & (
  | {
      mode: "create";
      workspaceId: string;
      type: TProjectStatusType;
      position: number;
      status?: never;
    }
  | { mode: "edit"; status: TProjectStatus; workspaceId?: never; type?: never; position?: never }
);

export function ProjectStatusDialog(props: TProps) {
  const [isOpen, setIsOpen] = useState(false);
  const title =
    props.mode === "create"
      ? `Add ${projectStatusTypeLabels[props.type].toLowerCase()} status`
      : "Edit status";
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger render={props.trigger} />
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        {isOpen && <ProjectStatusForm {...props} onDone={() => setIsOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function ProjectStatusForm(props: TProps & { onDone: () => void }) {
  const { mutateAsync: createStatus } = useCreateProjectStatusMutation();
  const { mutateAsync: updateStatus } = useUpdateProjectStatusMutation();
  const form = useAppForm({
    defaultValues:
      props.mode === "edit"
        ? {
            name: props.status.name,
            color: props.status.color,
            description: props.status.description,
          }
        : { name: "", color: defaultProjectStatusColors[props.type], description: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      if (props.mode === "edit") {
        await updateStatus({ status_id: props.status.id, ...parsed });
        props.onDone();
        return;
      }
      await createStatus({
        workspace_id: props.workspaceId,
        type: props.type,
        position: props.position,
        ...parsed,
      });
      props.onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField name="color" children={(field) => <field.ColorField label="Color" />} />
      <form.AppField name="name" children={(field) => <field.TextField label="Name" autoFocus />} />
      <form.AppField
        name="description"
        children={(field) => <field.TextareaField label="Description" rows={2} />}
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
            {props.mode === "edit" ? "Save" : "Add status"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
