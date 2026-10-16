import { FileIcon, LinkIcon, UploadIcon, XIcon } from "lucide-react";
import { type FormEvent, useState } from "react";

import {
  type TAttachment,
  type TIssueDetail,
  useCreateAttachmentMutation,
  useDeleteAttachmentMutation,
} from "@/lib/queries/issues";
import { List, ListRow, ListRowText } from "@/components/layout/list";
import { SectionTitle } from "@/components/layout/section-title";
import { RelativeTime } from "@/components/shared/relative-time";
import { Button } from "@/components/ui/button";
import { FileTrigger } from "@/components/ui/file-trigger";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { useCreateUploadMutation } from "@/lib/queries/uploads";

export function IssueAttachments({ issue }: { issue: TIssueDetail }) {
  const { workspace, canEdit } = useWorkspace();
  const attachments = issue.attachments ?? [];
  const { mutateAsync: upload, isPending: isUploading } = useCreateUploadMutation();
  const { mutateAsync: createAttachment } = useCreateAttachmentMutation();

  const uploadFiles = async (files: File[]) => {
    for (const file of files) {
      const res = await upload({ workspaceId: workspace.id, file });
      await createAttachment({ issueId: issue.id, title: file.name, uploadId: res.data.id });
    }
  };

  return (
    <section aria-label="Attachments" className="flex w-full flex-col gap-2">
      <SectionTitle
        title="Attachments"
        actions={
          canEdit && (
            <>
              <FileTrigger
                multiple
                onSelect={(files) => void uploadFiles(files).catch(() => undefined)}
              >
                {(open) => (
                  <Button variant="ghost" size="sm" isPending={isUploading} onClick={open}>
                    <UploadIcon className="size-4" />
                    Upload
                  </Button>
                )}
              </FileTrigger>
              <AddLinkPopover issueId={issue.id} />
            </>
          )
        }
      />
      {attachments.length > 0 && (
        <List>
          {attachments.map((attachment) => (
            <AttachmentRow key={attachment.id} attachment={attachment} canEdit={canEdit} />
          ))}
        </List>
      )}
    </section>
  );
}

function AttachmentRow({ attachment, canEdit }: { attachment: TAttachment; canEdit: boolean }) {
  const { mutate: remove, isPending } = useDeleteAttachmentMutation();
  const Icon = attachment.upload_id ? FileIcon : LinkIcon;
  return (
    <ListRow
      actions={
        <>
          <span className="text-muted-foreground text-xs">
            <RelativeTime value={attachment.created_at} />
          </span>
          {canEdit && (
            <Button
              variant="ghost"
              size="icon"
              aria-label={`Remove ${attachment.title}`}
              isPending={isPending}
              onClick={() => remove({ attachmentId: attachment.id })}
            >
              <XIcon className="size-4" />
            </Button>
          )}
        </>
      }
    >
      <Icon className="text-muted-foreground size-4 shrink-0" />
      <a
        href={attachment.url}
        target="_blank"
        rel="noreferrer noopener"
        className="flex min-w-0 flex-1 outline-hidden focus-visible:underline has-hover:hover:underline"
      >
        <ListRowText primary={attachment.title} secondary={attachment.subtitle || undefined} />
      </a>
    </ListRow>
  );
}

function AddLinkPopover({ issueId }: { issueId: string }) {
  const [isOpen, setIsOpen] = useState(false);
  const [url, setUrl] = useState("");
  const [title, setTitle] = useState("");
  const { mutateAsync: createAttachment, isPending } = useCreateAttachmentMutation();
  const isValid = /^https?:\/\/\S+$/i.test(url.trim());

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!isValid) return;
    const trimmed = url.trim();
    await createAttachment({ issueId, url: trimmed, title: title.trim() || trimmed });
    setUrl("");
    setTitle("");
    setIsOpen(false);
  };

  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger
        render={
          <Button variant="ghost" size="sm">
            <LinkIcon className="size-4" />
            Link
          </Button>
        }
      />
      <PopoverContent align="end" className="w-80">
        <form
          className="flex flex-col gap-4"
          onSubmit={(event) => void submit(event).catch(() => undefined)}
        >
          <Input
            aria-label="URL"
            placeholder="https://"
            autoFocus
            value={url}
            onChange={(event) => setUrl(event.target.value)}
          />
          <Input
            aria-label="Title"
            placeholder="Title (optional)"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
          />
          <Button
            type="submit"
            size="sm"
            isPending={isPending}
            disabled={!isValid}
            className="self-end"
          >
            Add link
          </Button>
        </form>
      </PopoverContent>
    </Popover>
  );
}
