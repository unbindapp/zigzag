import { AlarmClockIcon } from "lucide-react";

import { snoozeOptions } from "@/components/inbox/snooze";
import { useNow } from "@/components/providers/now-provider";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Kbd } from "@/components/ui/kbd";
import { formatDateTime } from "@/lib/helpers/dates";

type TProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSnooze: (until: Date) => void;
  disabled?: boolean;
};

export function SnoozeMenu({ open, onOpenChange, onSnooze, disabled }: TProps) {
  const now = useNow();
  const options = snoozeOptions(new Date(now));
  return (
    <DropdownMenu open={open} onOpenChange={onOpenChange}>
      <DropdownMenuTrigger
        render={
          <Button variant="ghost" size="icon" aria-label="Snooze" disabled={disabled}>
            <AlarmClockIcon className="size-4" />
          </Button>
        }
      />
      <DropdownMenuContent align="end" className="w-80">
        <DropdownMenuGroup>
          <DropdownMenuLabel className="flex items-center gap-2">
            <span className="flex-1">Snooze until</span>
            <Kbd>H</Kbd>
          </DropdownMenuLabel>
          {options.map((option) => (
            <DropdownMenuItem key={option.id} onClick={() => onSnooze(option.until)}>
              <span className="flex-1">{option.label}</span>
              <span className="text-muted-foreground text-xs">{formatDateTime(option.until)}</span>
            </DropdownMenuItem>
          ))}
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
