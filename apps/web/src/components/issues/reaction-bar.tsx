import { SmilePlusIcon } from "lucide-react";
import { useTheme } from "next-themes";
import { lazy, Suspense, useState } from "react";

import type { TReactionGroup } from "@/lib/queries/issues";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Skeleton } from "@/components/ui/skeleton";
import { SimpleTooltip } from "@/components/ui/tooltip";

const EmojiPicker = lazy(() => import("@/components/shared/emoji-picker"));

type TProps = {
  reactions: TReactionGroup[];
  onToggle: (emoji: string, reacted: boolean) => void;
  disabled?: boolean;
};

function reactorNames(reaction: TReactionGroup) {
  return reaction.users.map((user) => user.name || user.display_name).join(", ");
}

export function ReactionBar({ reactions, onToggle, disabled }: TProps) {
  const [isOpen, setIsOpen] = useState(false);
  const { resolvedTheme } = useTheme();
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {reactions.map((reaction) => (
        <SimpleTooltip key={reaction.emoji} content={reactorNames(reaction)}>
          <Button
            variant="outline"
            size="sm"
            forceMinSize={false}
            disabled={disabled}
            aria-pressed={reaction.reacted}
            aria-label={`${reaction.emoji} ${reaction.count}`}
            data-active={reaction.reacted || undefined}
            className="data-active:border-primary/6-10 data-active:bg-primary/1-10 rounded-full"
            onClick={() => onToggle(reaction.emoji, reaction.reacted)}
          >
            <span>{reaction.emoji}</span>
            <span className="text-xs">{reaction.count}</span>
          </Button>
        </SimpleTooltip>
      ))}
      {!disabled && (
        <Popover open={isOpen} onOpenChange={setIsOpen}>
          <PopoverTrigger
            render={
              <Button variant="ghost" size="icon" aria-label="Add reaction">
                <SmilePlusIcon className="text-muted-foreground size-4" />
              </Button>
            }
          />
          <PopoverContent align="start" className="w-80 p-2">
            <Suspense fallback={<Skeleton className="h-80 w-full" />}>
              <EmojiPicker
                theme={resolvedTheme === "light" ? "light" : "dark"}
                onSelect={(emoji) => {
                  setIsOpen(false);
                  const existing = reactions.find((reaction) => reaction.emoji === emoji);
                  onToggle(emoji, existing?.reacted ?? false);
                }}
              />
            </Suspense>
          </PopoverContent>
        </Popover>
      )}
    </div>
  );
}
