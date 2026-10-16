import { useRef, useState } from "react";

import { useUploadImage } from "@/components/issues/comment-thread";
import { useWorkspaceUsers } from "@/components/issues/picker-sources";
import { MarkdownEditor } from "@/components/shared/markdown-editor";

type TProps = {
  initialValue: string;
  label: string;
  placeholder: string;
  readOnly: boolean;
  onSave: (markdown: string) => Promise<unknown>;
};

export function AutosaveMarkdown({ initialValue, label, placeholder, readOnly, onSave }: TProps) {
  const [value] = useState(initialValue);
  const savedRef = useRef(initialValue);
  const { users } = useWorkspaceUsers();
  const uploadImage = useUploadImage();
  return (
    <MarkdownEditor
      label={label}
      value={value}
      placeholder={placeholder}
      readOnly={readOnly}
      users={users}
      onUploadImage={uploadImage}
      autosave={{
        onSave: async (markdown) => {
          if (markdown === savedRef.current) return;
          await onSave(markdown);
          savedRef.current = markdown;
        },
      }}
    />
  );
}
