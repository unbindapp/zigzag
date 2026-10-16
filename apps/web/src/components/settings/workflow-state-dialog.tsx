import { ReactElement, useState } from "react";
import { z } from "zod";

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
import { type TStateType, stateTypeLabels } from "@/components/shared/state-icon";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  type TWorkflowState,
  useCreateWorkflowStateMutation,
  useUpdateWorkflowStateMutation,
} from "@/lib/queries/teams";

export const defaultStateColors: Record<TStateType, string> = {
  triage: "#f2994a",
  backlog: "#95a2b3",
  unstarted: "#e2e2e2",
  started: "#f2c94c",
  completed: "#5e6ad2",
  canceled: "#95a2b3",
};

const formSchema = z.object({
  name: z.string().trim().min(1, "Name is required.").max(64, "Keep the name under 64 characters."),
  color: z.string(),
  description: z.string().max(255, "Keep the description under 255 characters."),
});

type TProps = {
  trigger: ReactElement;
} & (
  | { mode: "create"; teamId: string; type: TStateType; position: number; state?: never }
  | { mode: "edit"; state: TWorkflowState; teamId?: never; type?: never; position?: never }
);

export function WorkflowStateDialog(props: TProps) {
  const [isOpen, setIsOpen] = useState(false);
  const title =
    props.mode === "create"
      ? `Add ${stateTypeLabels[props.type].toLowerCase()} state`
      : "Edit state";
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger render={props.trigger} />
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        {isOpen && <WorkflowStateForm {...props} onDone={() => setIsOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function WorkflowStateForm(props: TProps & { onDone: () => void }) {
  const { mutateAsync: createState } = useCreateWorkflowStateMutation();
  const { mutateAsync: updateState } = useUpdateWorkflowStateMutation();
  const form = useAppForm({
    defaultValues:
      props.mode === "edit"
        ? {
            name: props.state.name,
            color: props.state.color,
            description: props.state.description ?? "",
          }
        : { name: "", color: defaultStateColors[props.type], description: "" },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      if (props.mode === "edit") {
        await updateState({ state_id: props.state.id, ...parsed });
        props.onDone();
        return;
      }
      await createState({
        team_id: props.teamId,
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
            {props.mode === "edit" ? "Save" : "Add state"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
