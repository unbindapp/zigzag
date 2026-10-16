import { CodeBlockLowlight } from "@tiptap/extension-code-block-lowlight";
import { Image } from "@tiptap/extension-image";
import { Link } from "@tiptap/extension-link";
import { Placeholder } from "@tiptap/extension-placeholder";
import { TaskItem } from "@tiptap/extension-task-item";
import { TaskList } from "@tiptap/extension-task-list";
import { Markdown } from "@tiptap/markdown";
import { type Editor, EditorContent, useEditor } from "@tiptap/react";
import { StarterKit } from "@tiptap/starter-kit";
import { common, createLowlight } from "lowlight";
import { useEffect, useMemo, useRef, useState } from "react";

import { Avatar } from "@/components/shared/avatar";
import { MarkdownMention } from "@/components/shared/markdown-mention";
import type { TUserSummary } from "@/components/shared/user-chip";
import { SuggestionMenu, type TSuggestionItem } from "@/components/ui/suggestion-menu";
import { toast } from "@/components/ui/toast";
import useDebouncedCallback from "@/lib/hooks/use-debounced-callback";

const lowlight = createLowlight(common);
const AUTOSAVE_DELAY_MS = 800;
const MAX_MENTION_RESULTS = 8;

type TSaveStatus = "idle" | "saving" | "saved" | "error";

type TMentionState = {
  items: TSuggestionItem[];
  index: number;
  rect: DOMRect | null;
  command: (item: TSuggestionItem) => void;
};

type TProps = {
  value: string;
  onChange?: (markdown: string) => void;
  placeholder?: string;
  users?: TUserSummary[];
  onUploadImage?: (file: File) => Promise<string>;
  readOnly?: boolean;
  autosave?: {
    onSave: (markdown: string) => Promise<unknown>;
    delayMs?: number;
  };
  onSubmit?: () => void;
  autoFocus?: boolean;
  label?: string;
};

export function MarkdownEditor({
  value,
  onChange,
  placeholder = "Write something…",
  users,
  onUploadImage,
  readOnly = false,
  autosave,
  onSubmit,
  autoFocus,
  label = "Editor",
}: TProps) {
  const usersRef = useRef(users ?? []);
  const uploadRef = useRef(onUploadImage);
  const submitRef = useRef(onSubmit);
  const changeRef = useRef(onChange);
  const editorRef = useRef<Editor | null>(null);
  const lastMarkdownRef = useRef(value);
  const [mention, setMention] = useState<TMentionState | null>(null);
  const mentionRef = useRef<TMentionState | null>(null);
  const [saveStatus, setSaveStatus] = useState<TSaveStatus>("idle");

  useEffect(() => {
    usersRef.current = users ?? [];
    uploadRef.current = onUploadImage;
    submitRef.current = onSubmit;
    changeRef.current = onChange;
  });

  const updateMention = (next: TMentionState | null) => {
    mentionRef.current = next;
    setMention(next);
  };

  const save = useDebouncedCallback(async (markdown: string) => {
    if (!autosave) return;
    setSaveStatus("saving");
    try {
      await autosave.onSave(markdown);
      setSaveStatus("saved");
    } catch (error) {
      setSaveStatus("error");
      toast.add({
        type: "error",
        title: "Could not save",
        description: error instanceof Error ? error.message : undefined,
      });
    }
  }, autosave?.delayMs ?? AUTOSAVE_DELAY_MS);

  useEffect(() => () => save.flush(), [save]);

  const extensions = useMemo(
    () => [
      StarterKit.configure({ codeBlock: false, link: false }),
      Link.configure({ openOnClick: false, autolink: true }),
      CodeBlockLowlight.configure({ lowlight }),
      TaskList,
      TaskItem.configure({ nested: true }),
      Image,
      Placeholder.configure({ placeholder }),
      Markdown,
      MarkdownMention.configure({
        HTMLAttributes: { class: "mention" },
        suggestion: {
          char: "@",
          items: ({ query }) => mentionItems(usersRef.current, query),
          render: () => ({
            onStart: (props) =>
              updateMention({
                items: props.items,
                index: 0,
                rect: props.clientRect?.() ?? null,
                command: (item) => props.command({ id: item.id, label: item.label }),
              }),
            onUpdate: (props) =>
              updateMention({
                items: props.items,
                index: 0,
                rect: props.clientRect?.() ?? null,
                command: (item) => props.command({ id: item.id, label: item.label }),
              }),
            onKeyDown: ({ event }) => handleMentionKey(event, mentionRef.current, updateMention),
            onExit: () => updateMention(null),
          }),
        },
      }),
    ],
    [placeholder],
  );

  const insertImages = async (files: File[]) => {
    const upload = uploadRef.current;
    const editor = editorRef.current;
    if (!upload || !editor) return;
    for (const file of files) {
      try {
        const src = await upload(file);
        editor.chain().focus().setImage({ src, alt: file.name }).run();
      } catch (error) {
        toast.add({
          type: "error",
          title: "Image upload failed",
          description: error instanceof Error ? error.message : undefined,
        });
      }
    }
  };

  const editor = useEditor({
    extensions,
    content: value,
    contentType: "markdown",
    editable: !readOnly,
    autofocus: autoFocus ? "end" : false,
    shouldRerenderOnTransaction: false,
    editorProps: {
      attributes: { "aria-label": label, class: "min-h-16 outline-none" },
      handleKeyDown: (_view, event) => {
        if (event.key !== "Enter" || !(event.metaKey || event.ctrlKey)) return false;
        if (!submitRef.current) return false;
        submitRef.current();
        return true;
      },
      handlePaste: (_view, event) => {
        const files = imageFiles(event.clipboardData?.files);
        if (files.length === 0 || !uploadRef.current) return false;
        void insertImages(files);
        return true;
      },
      handleDrop: (_view, event) => {
        const files = imageFiles(event.dataTransfer?.files);
        if (files.length === 0 || !uploadRef.current) return false;
        event.preventDefault();
        void insertImages(files);
        return true;
      },
    },
    onUpdate: ({ editor: updated }) => {
      const markdown = updated.getMarkdown();
      lastMarkdownRef.current = markdown;
      changeRef.current?.(markdown);
      if (autosave) save(markdown);
    },
  });

  useEffect(() => {
    editorRef.current = editor;
  }, [editor]);

  useEffect(() => {
    if (!editor || editor.isDestroyed) return;
    editor.setEditable(!readOnly);
  }, [editor, readOnly]);

  useEffect(() => {
    if (!editor || editor.isDestroyed) return;
    if (value === lastMarkdownRef.current) return;
    lastMarkdownRef.current = value;
    editor.commands.setContent(value, { contentType: "markdown", emitUpdate: false });
  }, [editor, value]);

  return (
    <div data-slot="markdown-editor" className="flex w-full flex-col gap-1.5">
      <EditorContent editor={editor} className="zz-prose w-full" />
      {mention && (
        <SuggestionMenu
          items={mention.items}
          selectedIndex={mention.index}
          rect={mention.rect}
          onSelect={mention.command}
          emptyText="No matching people"
        />
      )}
      {autosave && <SaveIndicator status={saveStatus} />}
    </div>
  );
}

function SaveIndicator({ status }: { status: TSaveStatus }) {
  const text = saveStatusText[status];
  return (
    <p aria-live="polite" className="text-muted-foreground h-4 self-end text-xs">
      {text}
    </p>
  );
}

const saveStatusText: Record<TSaveStatus, string> = {
  idle: "",
  saving: "Saving…",
  saved: "Saved",
  error: "Not saved",
};

function imageFiles(list: FileList | null | undefined) {
  return Array.from(list ?? []).filter((file) => file.type.startsWith("image/"));
}

function mentionItems(users: TUserSummary[], query: string): TSuggestionItem[] {
  const needle = query.toLowerCase();
  return users
    .filter(
      (user) =>
        user.name.toLowerCase().includes(needle) ||
        (user.display_name ?? "").toLowerCase().includes(needle),
    )
    .slice(0, MAX_MENTION_RESULTS)
    .map((user) => ({
      id: user.id,
      label: user.display_name || user.name,
      description: user.display_name ? user.name : undefined,
      icon: <Avatar id={user.id} name={user.name} src={user.avatar_url} size="xs" />,
    }));
}

function handleMentionKey(
  event: KeyboardEvent,
  state: TMentionState | null,
  update: (next: TMentionState | null) => void,
) {
  if (!state) return false;
  const count = state.items.length;
  switch (event.key) {
    case "ArrowDown":
      if (count === 0) return true;
      update({ ...state, index: (state.index + 1) % count });
      return true;
    case "ArrowUp":
      if (count === 0) return true;
      update({ ...state, index: (state.index - 1 + count) % count });
      return true;
    case "Enter":
    case "Tab": {
      const item = state.items[state.index];
      if (!item) return false;
      state.command(item);
      return true;
    }
    case "Escape":
      update(null);
      return true;
    default:
      return false;
  }
}
