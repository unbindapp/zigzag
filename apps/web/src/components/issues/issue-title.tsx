import { useEffect, useState } from "react";

import { Textarea } from "@/components/ui/textarea";

const TITLE_MAX_LENGTH = 255;

type TProps = {
  title: string;
  onSave: (title: string) => void;
  readOnly?: boolean;
};

export function IssueTitle({ title, onSave, readOnly }: TProps) {
  const [draft, setDraft] = useState(title);
  useEffect(() => setDraft(title), [title]);

  const commit = () => {
    const next = draft.replace(/\s+/g, " ").trim();
    if (!next) {
      setDraft(title);
      return;
    }
    if (next !== title) onSave(next);
  };

  return (
    <Textarea
      aria-label="Issue title"
      rows={1}
      maxLength={TITLE_MAX_LENGTH}
      readOnly={readOnly}
      value={draft}
      onChange={(event) => setDraft(event.target.value.replace(/\n/g, " "))}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          event.preventDefault();
          event.currentTarget.blur();
        }
        if (event.key === "Escape") {
          setDraft(title);
          event.currentTarget.blur();
        }
      }}
      className="field-sizing-content min-h-0 resize-none border-transparent bg-transparent px-0 py-0 text-xl leading-tight font-semibold focus-visible:ring-0 has-hover:hover:ring-0"
    />
  );
}
