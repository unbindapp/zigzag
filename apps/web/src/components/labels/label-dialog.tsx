import { ReactElement, useState } from "react";

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
import { colorFromId } from "@/lib/helpers/colors";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  LabelFormSchema,
  type TLabel,
  type TLabelTree,
  useCreateLabelMutation,
  useUpdateLabelMutation,
} from "@/lib/queries/labels";

const NO_GROUP = "none";

type TProps = {
  trigger?: ReactElement;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  groups: TLabelTree[];
} & (
  | {
      mode: "create";
      workspaceId: string;
      teamId?: string;
      isGroup: boolean;
      parentId?: string;
      label?: never;
    }
  | {
      mode: "edit";
      label: TLabel;
      workspaceId?: never;
      teamId?: never;
      isGroup?: never;
      parentId?: never;
    }
);

export function LabelDialog(props: TProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = props.open ?? internalOpen;
  const setOpen = (next: boolean) => {
    setInternalOpen(next);
    props.onOpenChange?.(next);
  };
  const isGroup = props.mode === "edit" ? props.label.is_group : props.isGroup;
  const noun = isGroup ? "group" : "label";
  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      {props.trigger && <DialogTrigger render={props.trigger} />}
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{props.mode === "edit" ? `Edit ${noun}` : `New ${noun}`}</DialogTitle>
        </DialogHeader>
        {isOpen && <LabelForm {...props} asGroup={isGroup} onDone={() => setOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function LabelForm({
  asGroup,
  onDone,
  ...props
}: TProps & { asGroup: boolean; onDone: () => void }) {
  const { mutateAsync: createLabel } = useCreateLabelMutation();
  const { mutateAsync: updateLabel } = useUpdateLabelMutation();
  const groupItems = [
    { value: NO_GROUP, label: "No group" },
    ...props.groups.map((group) => ({ value: group.id, label: group.name })),
  ];
  const form = useAppForm({
    defaultValues:
      props.mode === "edit"
        ? {
            name: props.label.name,
            color: props.label.color,
            description: props.label.description,
            parentId: props.label.parent_id ?? NO_GROUP,
            groupSingleSelect: props.label.group_single_select,
          }
        : {
            name: "",
            color: colorFromId(crypto.randomUUID()),
            description: "",
            parentId: props.parentId ?? NO_GROUP,
            groupSingleSelect: false,
          },
    validators: { onSubmit: LabelFormSchema },
    onSubmit: async ({ value }) => {
      const parsed = LabelFormSchema.parse(value);
      const parentId = parsed.parentId === NO_GROUP ? undefined : parsed.parentId;
      if (props.mode === "edit") {
        await updateLabel({
          label_id: props.label.id,
          name: parsed.name,
          color: parsed.color,
          description: parsed.description,
          ...(asGroup
            ? { group_single_select: parsed.groupSingleSelect }
            : { parent_id: parentId ?? "00000000-0000-0000-0000-000000000000" }),
        });
        onDone();
        return;
      }
      await createLabel({
        workspace_id: props.workspaceId,
        team_id: props.teamId,
        name: parsed.name,
        color: parsed.color,
        description: parsed.description || undefined,
        is_group: asGroup || undefined,
        group_single_select: asGroup ? parsed.groupSingleSelect : undefined,
        parent_id: asGroup ? undefined : parentId,
      });
      onDone();
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
      {asGroup ? (
        <form.AppField
          name="groupSingleSelect"
          children={(field) => (
            <field.SwitchField
              label="Single select"
              description="An issue can carry only one label of this group."
            />
          )}
        />
      ) : (
        groupItems.length > 1 && (
          <form.AppField
            name="parentId"
            children={(field) => <field.SelectField label="Group" items={groupItems} />}
          />
        )
      )}
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
            {props.mode === "edit" ? "Save" : asGroup ? "Create group" : "Create label"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
