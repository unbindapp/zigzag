import * as React from "react";
import { createPortal } from "react-dom";

import { Button } from "@/components/ui/button";
import { cn } from "@/components/ui/utils";

export type TSuggestionItem = {
  id: string;
  label: string;
  description?: string;
  icon?: React.ReactNode;
};

type TProps = {
  items: TSuggestionItem[];
  selectedIndex: number;
  rect: DOMRect | null;
  onSelect: (item: TSuggestionItem) => void;
  emptyText?: string;
};

function SuggestionMenu({
  items,
  selectedIndex,
  rect,
  onSelect,
  emptyText = "No matches",
}: TProps) {
  if (!rect) return null;
  return createPortal(
    <div
      role="listbox"
      data-slot="suggestion-menu"
      className="bg-popover text-popover-foreground shadow-shadow-color/shadow-opacity fixed z-[1100] flex max-h-64 w-64 flex-col overflow-y-auto rounded-lg border p-1 shadow-lg"
      style={{ top: rect.bottom + 4, left: rect.left }}
    >
      {items.length === 0 && (
        <p className="text-muted-foreground px-2 py-1.5 text-sm">{emptyText}</p>
      )}
      {items.map((item, index) => (
        <Button
          key={item.id}
          type="button"
          role="option"
          aria-selected={index === selectedIndex}
          variant="ghost"
          size="sm"
          forceMinSize={false}
          onMouseDown={(event) => event.preventDefault()}
          onClick={() => onSelect(item)}
          className={cn(
            "w-full justify-start font-normal",
            index === selectedIndex && "bg-accent text-accent-foreground",
          )}
        >
          {item.icon}
          <span className="truncate">{item.label}</span>
          {item.description && (
            <span className="text-muted-foreground ml-auto truncate text-xs">
              {item.description}
            </span>
          )}
        </Button>
      ))}
    </div>,
    document.body,
  );
}

export { SuggestionMenu };
