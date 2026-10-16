import { HammerIcon } from "lucide-react";

import { EmptyState } from "@/components/layout/empty-state";
import { useNewIssueStore } from "@/components/issues/new-issue-store";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";

export function NewIssueDialog() {
  const isOpen = useNewIssueStore((state) => state.isOpen);
  const close = useNewIssueStore((state) => state.close);
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && close()}>
      <DialogContent className="w-xl max-w-full">
        <DialogTitle>New issue</DialogTitle>
        <EmptyState icon={HammerIcon} title="Not built yet" />
      </DialogContent>
    </Dialog>
  );
}
