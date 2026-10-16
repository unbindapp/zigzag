import { ReactElement, ReactNode, useId, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

type TProps = {
  trigger?: ReactElement;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  title: string;
  description?: ReactNode;
  confirmLabel: string;
  confirmText?: string;
  variant?: "destructive" | "default";
  onConfirm: () => Promise<unknown>;
  children?: ReactNode;
  confirmDisabled?: boolean;
};

export function ConfirmDialog({
  trigger,
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  confirmText,
  variant = "destructive",
  onConfirm,
  children,
  confirmDisabled,
}: TProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const [typed, setTyped] = useState("");
  const [isPending, setIsPending] = useState(false);
  const inputId = useId();
  const isOpen = open ?? internalOpen;
  const setOpen = (next: boolean) => {
    if (!next) setTyped("");
    setInternalOpen(next);
    onOpenChange?.(next);
  };
  const isConfirmed = !confirmText || typed.trim() === confirmText;

  const confirm = async () => {
    setIsPending(true);
    try {
      await onConfirm();
      setOpen(false);
    } catch {
      return;
    } finally {
      setIsPending(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setOpen}>
      {trigger && <DialogTrigger render={trigger} />}
      <DialogContent className="w-md max-w-full">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description && <DialogDescription>{description}</DialogDescription>}
        </DialogHeader>
        {children}
        {confirmText && (
          <div className="flex flex-col gap-1.5">
            <label htmlFor={inputId} className="text-sm">
              Type <span className="font-mono text-xs font-medium">{confirmText}</span> to confirm
            </label>
            <Input
              id={inputId}
              value={typed}
              autoComplete="off"
              onChange={(event) => setTyped(event.target.value)}
            />
          </div>
        )}
        <DialogFooter>
          <DialogClose
            render={
              <Button variant="outline" size="sm">
                Cancel
              </Button>
            }
          />
          <Button
            variant={variant}
            size="sm"
            isPending={isPending}
            disabled={!isConfirmed || confirmDisabled}
            onClick={confirm}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
