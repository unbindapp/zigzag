import { useEffect, useState } from "react";

import { Textarea } from "@/components/ui/textarea";

type TProps = {
  value: string;
  onSave: (value: string) => void;
  label: string;
  maxLength: number;
  readOnly?: boolean;
  placeholder?: string;
  allowEmpty?: boolean;
  variant?: "title" | "summary";
};

export function InlineText({
  value,
  onSave,
  label,
  maxLength,
  readOnly,
  placeholder,
  allowEmpty = false,
  variant = "title",
}: TProps) {
  const [draft, setDraft] = useState(value);
  useEffect(() => setDraft(value), [value]);

  const commit = () => {
    const next = draft.replace(/\s+/g, " ").trim();
    if (!next && !allowEmpty) {
      setDraft(value);
      return;
    }
    if (next !== value) onSave(next);
  };

  return (
    <Textarea
      aria-label={label}
      rows={1}
      maxLength={maxLength}
      readOnly={readOnly}
      placeholder={placeholder}
      value={draft}
      onChange={(event) => setDraft(event.target.value.replace(/\n/g, " "))}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          event.preventDefault();
          event.currentTarget.blur();
        }
        if (event.key === "Escape") {
          setDraft(value);
          event.currentTarget.blur();
        }
      }}
      data-variant={variant}
      className="text-muted-foreground data-[variant=title]:text-foreground field-sizing-content min-h-0 resize-none border-transparent bg-transparent px-0 py-0 text-sm focus-visible:ring-0 has-hover:hover:ring-0 data-[variant=title]:text-xl data-[variant=title]:leading-tight data-[variant=title]:font-semibold"
    />
  );
}
