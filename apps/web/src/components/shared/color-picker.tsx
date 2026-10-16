import { CheckIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { isHexColor, PRESET_COLORS } from "@/lib/helpers/colors";

type TProps = {
  value: string | null | undefined;
  onChange: (color: string) => void;
  disabled?: boolean;
  label?: string;
};

export function ColorPicker({ value, onChange, disabled, label = "Color" }: TProps) {
  const [isOpen, setIsOpen] = useState(false);
  const current = isHexColor(value) ? value : null;
  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger
        disabled={disabled}
        render={
          <Button variant="outline" size="sm" aria-label={`${label}: ${current ?? "default"}`}>
            <span
              aria-hidden
              data-empty={current ? undefined : true}
              className="data-empty:border-muted-foreground size-3.5 shrink-0 rounded-full data-empty:border data-empty:border-dashed"
              style={current ? { backgroundColor: current } : undefined}
            />
            <span className="font-mono text-xs">{current ?? "Default"}</span>
          </Button>
        }
      />
      <PopoverContent align="start" className="flex w-60 flex-col gap-3 p-3">
        <ColorSwatches
          value={current}
          onChange={(color) => {
            onChange(color);
            setIsOpen(false);
          }}
        />
        <HexInput value={current ?? ""} onChange={onChange} />
      </PopoverContent>
    </Popover>
  );
}

function ColorSwatches({
  value,
  onChange,
}: {
  value: string | null;
  onChange: (color: string) => void;
}) {
  return (
    <div role="listbox" aria-label="Preset colors" className="grid grid-cols-6 gap-1">
      {PRESET_COLORS.map((color) => (
        <Button
          key={color}
          type="button"
          role="option"
          aria-selected={color === value}
          aria-label={color}
          variant="ghost"
          size="icon"
          forceMinSize={false}
          onClick={() => onChange(color)}
        >
          <span
            aria-hidden
            className="flex size-5 items-center justify-center rounded-full text-white"
            style={{ backgroundColor: color }}
          >
            {color.toLowerCase() === value?.toLowerCase() && <CheckIcon className="size-3.5" />}
          </span>
        </Button>
      ))}
    </div>
  );
}

function HexInput({ value, onChange }: { value: string; onChange: (color: string) => void }) {
  const [draft, setDraft] = useState(value);
  const commit = () => {
    const normalized = draft.startsWith("#") ? draft : `#${draft}`;
    if (!isHexColor(normalized)) {
      setDraft(value);
      return;
    }
    onChange(normalized.toLowerCase());
  };
  return (
    <Input
      aria-label="Hex color"
      placeholder="#5e6ad2"
      value={draft}
      maxLength={7}
      className="font-mono"
      onChange={(event) => setDraft(event.target.value)}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key !== "Enter") return;
        event.preventDefault();
        commit();
      }}
    />
  );
}
