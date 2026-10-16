import { TriangleAlertIcon } from "lucide-react";

import { GithubWebhookSteps } from "@/components/integrations/github-webhook-steps";
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
import type { TGithubSecret } from "@/lib/queries/integrations";

type TProps = {
  secret: TGithubSecret | null;
  onClose: () => void;
};

export function GithubSecretDialog({ secret, onClose }: TProps) {
  return (
    <Dialog open={secret !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent hideXButton className="w-lg max-w-full">
        <DialogHeader>
          <DialogTitle>GitHub webhook secret</DialogTitle>
          <DialogDescription>
            Add a webhook on GitHub with this URL and secret. The secret will not be shown again.
          </DialogDescription>
        </DialogHeader>
        <CopyField value={secret?.webhook_url ?? ""} label="Webhook URL" />
        <CopyField value={secret?.secret ?? ""} label="Secret" />
        <GithubWebhookSteps />
        <div className="bg-warning/3-10 text-warning flex items-start gap-1.5 rounded-md px-3 py-2 text-sm">
          <TriangleAlertIcon className="mt-0.5 size-4 shrink-0" />
          <p>Copy the secret now. Regenerate it if you lose it.</p>
        </div>
        <DialogFooter>
          <DialogClose render={<Button size="sm">Done</Button>} />
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
