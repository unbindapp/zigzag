import { SmileIcon, XIcon } from "lucide-react";
import { lazy, Suspense, useState } from "react";
import { useTheme } from "next-themes";

import { EntityIcon } from "@/components/shared/entity-icon";
import { ENTITY_ICON_NAMES, ENTITY_ICONS } from "@/components/shared/entity-icons";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsPanel, TabsTab } from "@/components/ui/tabs";

const EmojiPicker = lazy(() => import("@/components/shared/emoji-picker"));

type TProps = {
  value: string | null | undefined;
  onChange: (icon: string | null) => void;
  color?: string | null;
  disabled?: boolean;
  label?: string;
};

export function IconPicker({ value, onChange, color, disabled, label = "Icon" }: TProps) {
  const [isOpen, setIsOpen] = useState(false);
  const select = (icon: string | null) => {
    onChange(icon);
    setIsOpen(false);
  };
  return (
    <Popover open={isOpen} onOpenChange={setIsOpen}>
      <PopoverTrigger
        disabled={disabled}
        render={
          <Button variant="outline" size="icon" aria-label={label}>
            <EntityIcon icon={value} color={color} fallback={SmileIcon} />
          </Button>
        }
      />
      <PopoverContent align="start" className="flex w-80 flex-col gap-2 p-2">
        <Tabs defaultValue="icons" className="gap-2">
          <TabsList>
            <TabsTab value="icons">Icons</TabsTab>
            <TabsTab value="emoji">Emoji</TabsTab>
          </TabsList>
          <TabsPanel value="icons">
            <IconGrid value={value} color={color} onSelect={select} />
          </TabsPanel>
          <TabsPanel value="emoji">
            <Suspense fallback={<Skeleton className="h-80 w-full" />}>
              <EmojiPickerPanel onSelect={select} />
            </Suspense>
          </TabsPanel>
        </Tabs>
        {value && (
          <Button variant="ghost" size="sm" onClick={() => select(null)} className="self-start">
            <XIcon className="size-4" />
            Remove icon
          </Button>
        )}
      </PopoverContent>
    </Popover>
  );
}

function IconGrid({
  value,
  color,
  onSelect,
}: {
  value: string | null | undefined;
  color?: string | null;
  onSelect: (icon: string) => void;
}) {
  return (
    <div role="listbox" aria-label="Icons" className="grid grid-cols-8 gap-1">
      {ENTITY_ICON_NAMES.map((name) => {
        const Icon = ENTITY_ICONS[name];
        return (
          <Button
            key={name}
            type="button"
            role="option"
            aria-selected={name === value}
            aria-label={name}
            variant="ghost"
            size="icon"
            forceMinSize={false}
            data-selected={name === value || undefined}
            className="data-selected:bg-accent"
            onClick={() => onSelect(name)}
          >
            <Icon className="size-4" style={color ? { color } : undefined} />
          </Button>
        );
      })}
    </div>
  );
}

function EmojiPickerPanel({ onSelect }: { onSelect: (emoji: string) => void }) {
  const { resolvedTheme } = useTheme();
  return <EmojiPicker theme={resolvedTheme === "light" ? "light" : "dark"} onSelect={onSelect} />;
}
