import { CheckIcon, CopyIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useCopyToClipboard } from "@/lib/hooks/use-copy";

type TProps = {
  value: string | undefined;
  label?: string;
};

export function CopyButton({ value, label = "Copy" }: TProps) {
  const { copyToClipboard, isRecentlyCopied } = useCopyToClipboard();
  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      aria-label={isRecentlyCopied ? "Copied" : label}
      disabled={!value}
      onClick={() => value && copyToClipboard(value)}
    >
      {isRecentlyCopied ? <CheckIcon className="size-4" /> : <CopyIcon className="size-4" />}
    </Button>
  );
}

export function CopyField({ value, label }: { value: string; label: string }) {
  return (
    <div className="flex w-full items-center gap-2">
      <p
        aria-label={label}
        className="bg-input min-w-0 flex-1 truncate rounded-md border px-3 py-2 font-mono text-xs"
      >
        {value}
      </p>
      <CopyButton value={value} label={`Copy ${label.toLowerCase()}`} />
    </div>
  );
}
