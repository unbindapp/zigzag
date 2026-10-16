import { type ReactElement, useState } from "react";
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
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  type TMilestone,
  useCreateMilestoneMutation,
  useUpdateMilestoneMutation,
} from "@/lib/queries/projects";

const milestoneSchema = z.object({
  name: z.string().trim().min(1, "Name is required.").max(80, "Keep the name under 80 characters."),
  description: z.string().trim().max(5000, "Keep the description under 5000 characters."),
  target_date: z.string().regex(/^(\d{4}-\d{2}-\d{2})?$/, "Pick a date."),
});

type TProps = { trigger: ReactElement } & (
  | { mode: "create"; projectId: string; milestone?: never }
  | { mode: "edit"; milestone: TMilestone; projectId?: never }
);

export function MilestoneDialog(props: TProps) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger render={props.trigger} />
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{props.mode === "create" ? "Add milestone" : "Edit milestone"}</DialogTitle>
        </DialogHeader>
        {isOpen && <MilestoneForm {...props} onDone={() => setIsOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function MilestoneForm(props: TProps & { onDone: () => void }) {
  const { mutateAsync: createMilestone } = useCreateMilestoneMutation();
  const { mutateAsync: updateMilestone } = useUpdateMilestoneMutation();
  const form = useAppForm({
    defaultValues:
      props.mode === "edit"
        ? {
            name: props.milestone.name,
            description: props.milestone.description,
            target_date: props.milestone.target_date ?? "",
          }
        : { name: "", description: "", target_date: "" },
    validators: { onSubmit: milestoneSchema },
    onSubmit: async ({ value }) => {
      const parsed = milestoneSchema.parse(value);
      if (props.mode === "edit") {
        await updateMilestone({ milestone_id: props.milestone.id, ...parsed });
        props.onDone();
        return;
      }
      await createMilestone({
        project_id: props.projectId,
        name: parsed.name,
        description: parsed.description || undefined,
        target_date: parsed.target_date || undefined,
      });
      props.onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField name="name" children={(field) => <field.TextField label="Name" autoFocus />} />
      <form.AppField
        name="description"
        children={(field) => <field.TextareaField label="Description" rows={2} />}
      />
      <form.AppField
        name="target_date"
        children={(field) => <field.TextField label="Target date" type="date" />}
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
            {props.mode === "edit" ? "Save" : "Add milestone"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
