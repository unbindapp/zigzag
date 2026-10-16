import { ExternalLinkIcon, LinkIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";
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
  type TProject,
  useCreateProjectLinkMutation,
  useDeleteProjectLinkMutation,
} from "@/lib/queries/projects";

const linkSchema = z.object({
  url: z
    .string()
    .trim()
    .url("Enter a full URL such as https://example.com.")
    .refine((value) => /^https?:\/\//i.test(value), "Use an http or https URL."),
  label: z
    .string()
    .trim()
    .min(1, "Label is required.")
    .max(120, "Keep the label under 120 characters."),
});

export function ProjectLinks({ project, canEdit }: { project: TProject; canEdit: boolean }) {
  const { mutate: deleteLink } = useDeleteProjectLinkMutation();
  return (
    <div className="flex w-full min-w-0 flex-col gap-1.5">
      {project.links.map((link) => (
        <div key={link.id} className="group/link flex min-h-8 min-w-0 items-center gap-1.5 text-sm">
          <LinkIcon className="text-muted-foreground size-4 shrink-0" />
          <a
            href={link.url}
            target="_blank"
            rel="noreferrer"
            className="min-w-0 flex-1 truncate hover:underline"
          >
            {link.label}
          </a>
          <ExternalLinkIcon className="text-muted-foreground size-3.5 shrink-0" />
          {canEdit && (
            <Button
              variant="ghost"
              size="icon"
              aria-label={`Remove link ${link.label}`}
              onClick={() => deleteLink({ linkId: link.id })}
            >
              <Trash2Icon className="size-4" />
            </Button>
          )}
        </div>
      ))}
      {project.links.length === 0 && (
        <span className="text-muted-foreground text-sm">No links</span>
      )}
      {canEdit && <AddLinkDialog projectId={project.id} />}
    </div>
  );
}

function AddLinkDialog({ projectId }: { projectId: string }) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger
        render={
          <Button variant="ghost" size="sm" className="self-start">
            <PlusIcon className="size-4" />
            Add link
          </Button>
        }
      />
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>Add link</DialogTitle>
        </DialogHeader>
        {isOpen && <AddLinkForm projectId={projectId} onDone={() => setIsOpen(false)} />}
      </DialogContent>
    </Dialog>
  );
}

function AddLinkForm({ projectId, onDone }: { projectId: string; onDone: () => void }) {
  const { mutateAsync: createLink } = useCreateProjectLinkMutation();
  const form = useAppForm({
    defaultValues: { url: "", label: "" },
    validators: { onSubmit: linkSchema },
    onSubmit: async ({ value }) => {
      const parsed = linkSchema.parse(value);
      await createLink({ project_id: projectId, ...parsed });
      onDone();
    },
  });
  return (
    <form className="flex flex-col gap-4" onSubmit={submitHandler(form)} noValidate>
      <form.AppField
        name="url"
        children={(field) => <field.TextField label="URL" placeholder="https://" autoFocus />}
      />
      <form.AppField name="label" children={(field) => <field.TextField label="Label" />} />
      <DialogFooter>
        <DialogClose
          render={
            <Button type="button" variant="outline" size="sm">
              Cancel
            </Button>
          }
        />
        <form.AppForm>
          <form.SubmitButton size="sm">Add link</form.SubmitButton>
        </form.AppForm>
      </DialogFooter>
    </form>
  );
}
