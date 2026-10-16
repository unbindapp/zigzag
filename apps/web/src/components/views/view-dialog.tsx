import { ReactElement, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  type TView,
  type TViewDisplay,
  useCreateViewMutation,
  useUpdateViewMutation,
  ViewFormSchema,
} from "@/lib/queries/views";

export type TNewView = {
  workspaceId: string;
  teamId?: string;
  filters: Record<string, unknown>;
  display: TViewDisplay;
};

type TProps = {
  trigger?: ReactElement;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSaved?: (view: TView) => void;
} & (
  { mode: "create"; draft: TNewView; view?: never } | { mode: "edit"; view: TView; draft?: never }
);

export function ViewDialog(props: TProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = props.open ?? internalOpen;
  const setOpen = (next: boolean) => {
    setInternalOpen(next);
    props.onOpenChange?.(next);
  };
  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      {props.trigger && <DialogTrigger render={props.trigger} />}
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{props.mode === "edit" ? "Edit view" : "Save view"}</DialogTitle>
          {props.mode === "create" && (
            <DialogDescription>Saves the current filters and display options.</DialogDescription>
          )}
        </DialogHeader>
        {isOpen && <ViewForm {...props} onDone={() => setOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function ViewForm(props: TProps & { onDone: () => void }) {
  const { mutateAsync: createView } = useCreateViewMutation();
  const { mutateAsync: updateView } = useUpdateViewMutation();
  const form = useAppForm({
    defaultValues:
      props.mode === "edit"
        ? {
            name: props.view.name,
            description: props.view.description,
            icon: props.view.icon,
            color: props.view.color,
            shared: props.view.shared,
          }
        : { name: "", description: "", icon: "", color: "", shared: false },
    validators: { onSubmit: ViewFormSchema },
    onSubmit: async ({ value }) => {
      const parsed = ViewFormSchema.parse(value);
      const result =
        props.mode === "edit"
          ? await updateView({ view_id: props.view.id, ...parsed })
          : await createView({
              workspace_id: props.draft.workspaceId,
              team_id: props.draft.teamId,
              kind: "issue",
              filters: props.draft.filters,
              display: props.draft.display,
              name: parsed.name,
              description: parsed.description || undefined,
              icon: parsed.icon || undefined,
              color: parsed.color || undefined,
              shared: parsed.shared,
            });
      props.onSaved?.(result.data);
      props.onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <div className="flex items-start gap-2">
        <form.AppField name="icon" children={(field) => <field.IconField label="Icon" />} />
        <form.AppField name="color" children={(field) => <field.ColorField label="Color" />} />
      </div>
      <form.AppField name="name" children={(field) => <field.TextField label="Name" autoFocus />} />
      <form.AppField
        name="description"
        children={(field) => <field.TextareaField label="Description" rows={2} />}
      />
      <form.AppField
        name="shared"
        children={(field) => (
          <field.SwitchField
            label="Share with workspace"
            description="Everyone in the workspace can open shared views."
          />
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
            {props.mode === "edit" ? "Save" : "Save view"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
