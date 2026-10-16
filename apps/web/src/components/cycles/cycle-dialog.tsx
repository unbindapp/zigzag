import { format } from "date-fns";
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
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import { type TCycle, useCreateCycleMutation, useUpdateCycleMutation } from "@/lib/queries/cycles";
import type { TTeam } from "@/lib/queries/teams";

const datePattern = /^\d{4}-\d{2}-\d{2}$/;

const formSchema = z
  .object({
    name: z.string().trim().max(64, "Keep the name under 64 characters."),
    description: z.string().trim().max(2000, "Keep the description under 2000 characters."),
    startsOn: z.string().regex(datePattern, "Pick a start date."),
    endsOn: z.string().regex(datePattern, "Pick an end date."),
  })
  .refine((value) => value.endsOn >= value.startsOn, {
    message: "The cycle must end on or after its start.",
    path: ["endsOn"],
  });

function toDay(value: string | Date) {
  return format(new Date(value), "yyyy-MM-dd");
}

function startOfDay(day: string) {
  return new Date(`${day}T00:00:00`).toISOString();
}

function endOfDay(day: string) {
  const next = new Date(`${day}T00:00:00`);
  next.setDate(next.getDate() + 1);
  return next.toISOString();
}

type TProps = {
  trigger?: ReactElement;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
} & ({ mode: "create"; team: TTeam; after?: TCycle } | { mode: "edit"; cycle: TCycle });

export function CycleDialog(props: TProps) {
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
          <DialogTitle>{props.mode === "edit" ? "Edit cycle" : "New cycle"}</DialogTitle>
        </DialogHeader>
        {isOpen && <CycleForm {...props} onDone={() => setOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function defaultRange(team: TTeam, after?: TCycle) {
  const start = after ? new Date(after.ends_at) : new Date();
  const end = new Date(start);
  end.setDate(end.getDate() + 7 * team.cycle_duration_weeks - 1);
  return { startsOn: toDay(start), endsOn: toDay(end) };
}

function CycleForm(props: TProps & { onDone: () => void }) {
  const { mutateAsync: createCycle } = useCreateCycleMutation();
  const { mutateAsync: updateCycle } = useUpdateCycleMutation();
  const form = useAppForm({
    defaultValues:
      props.mode === "edit"
        ? {
            name: props.cycle.name,
            description: props.cycle.description,
            startsOn: toDay(props.cycle.starts_at),
            endsOn: toDay(new Date(new Date(props.cycle.ends_at).getTime() - 1)),
          }
        : { name: "", description: "", ...defaultRange(props.team, props.after) },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.parse(value);
      const dates = { starts_at: startOfDay(parsed.startsOn), ends_at: endOfDay(parsed.endsOn) };
      if (props.mode === "edit") {
        await updateCycle({
          cycle_id: props.cycle.id,
          name: parsed.name,
          description: parsed.description,
          ...dates,
        });
        props.onDone();
        return;
      }
      await createCycle({
        team_id: props.team.id,
        name: parsed.name || undefined,
        description: parsed.description || undefined,
        ...dates,
      });
      props.onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="name"
        children={(field) => <field.TextField label="Name" placeholder="Optional" autoFocus />}
      />
      <div className="grid grid-cols-2 gap-4">
        <form.AppField
          name="startsOn"
          children={(field) => <field.TextField label="Starts" type="date" />}
        />
        <form.AppField
          name="endsOn"
          children={(field) => <field.TextField label="Ends" type="date" />}
        />
      </div>
      <form.AppField
        name="description"
        children={(field) => <field.TextareaField label="Description" rows={3} />}
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
            {props.mode === "edit" ? "Save" : "Create cycle"}
          </form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
