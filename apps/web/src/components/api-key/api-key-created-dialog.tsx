import { TriangleAlertIcon } from "lucide-react";

import { CopyField } from "@/components/shared/copy-button";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { TApiKeyCreated } from "@/lib/queries/api-keys";

type TProps = {
  created: TApiKeyCreated | null;
  onClose: () => void;
};

export function ApiKeyCreatedDialog({ created, onClose }: TProps) {
  return (
    <Dialog open={created !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent hideXButton className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>API key created</DialogTitle>
          <DialogDescription>
            {created ? `${created.name} is ready. ` : ""}Keep it safe, it will not be shown again.
          </DialogDescription>
        </DialogHeader>
        <CopyField value={created?.token ?? ""} label="API key" />
        <div className="bg-warning/3-10 text-warning flex items-start gap-1.5 rounded-md px-3 py-2 text-sm">
          <TriangleAlertIcon className="mt-0.5 size-4 shrink-0" />
          <p>Copy the key now. Zigzag only stores a hash of it.</p>
        </div>
        <DialogFooter>
          <DialogClose render={<Button size="sm">Done</Button>} />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
