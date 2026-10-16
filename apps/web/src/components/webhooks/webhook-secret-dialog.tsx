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
import type { TWebhookCreated } from "@/lib/queries/webhooks";

type TProps = {
  created: TWebhookCreated | null;
  onClose: () => void;
};

export function WebhookSecretDialog({ created, onClose }: TProps) {
  return (
    <Dialog open={created !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent hideXButton className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>Webhook created</DialogTitle>
          <DialogDescription>
            Use this secret to check that deliveries come from Zigzag. It will not be shown again.
          </DialogDescription>
        </DialogHeader>
        <CopyField value={created?.secret ?? ""} label="Signing secret" />
        <p className="text-muted-foreground text-sm">
          Every delivery carries an X-Zigzag-Signature header: sha256= followed by the hex
          HMAC-SHA256 of the raw body, keyed with this secret.
        </p>
        <div className="bg-warning/3-10 text-warning flex items-start gap-1.5 rounded-md px-3 py-2 text-sm">
          <TriangleAlertIcon className="mt-0.5 size-4 shrink-0" />
          <p>Copy the secret now. Delete and recreate the webhook if you lose it.</p>
        </div>
        <DialogFooter>
          <DialogClose render={<Button size="sm">Done</Button>} />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
