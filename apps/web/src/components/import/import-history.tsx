import { useQuery } from "@tanstack/react-query";
import { HistoryIcon } from "lucide-react";
import { useState } from "react";

import { ImportReportView } from "@/components/import/import-report";
import { reportSummary } from "@/components/import/parse-export";
import { List, ListRow, ListRowText } from "@/components/layout/list";
import { ListStates } from "@/components/layout/list-states";
import { RelativeTime } from "@/components/shared/relative-time";
import { UserChip } from "@/components/shared/user-chip";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { importsListQuery, type TImport } from "@/lib/queries/imports";

const statusBadges: Record<
  TImport["status"],
  { label: string; variant: "process" | "success" | "destructive" }
> = {
  running: { label: "Running", variant: "process" },
  completed: { label: "Completed", variant: "success" },
  failed: { label: "Failed", variant: "destructive" },
};

export function ImportHistory() {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(importsListQuery(workspace.id));
  const [opened, setOpened] = useState<TImport | null>(null);
  const imports = data?.imports ?? [];
  return (
    <>
      <ListStates
        isPending={isPending}
        error={error}
        onRetry={refetch}
        isEmpty={imports.length === 0}
        skeletonRows={3}
        empty={{
          icon: HistoryIcon,
          title: "No imports yet",
          description: "Dry runs and imports of this workspace show up here with their reports.",
        }}
      >
        <List>
          {imports.map((item) => (
            <ImportRow key={item.id} item={item} onOpen={() => setOpened(item)} />
          ))}
        </List>
      </ListStates>
      <Dialog open={opened !== null} onOpenChange={(open) => !open && setOpened(null)}>
        <DialogContent className="w-2xl max-w-full">
          <DialogHeader>
            <DialogTitle>Import report</DialogTitle>
            <DialogDescription>
              {opened ? reportSummary(opened.report) : "The report of a past import."}
            </DialogDescription>
          </DialogHeader>
          {opened && <ImportReportView report={opened.report} />}
        </DialogContent>
      </Dialog>
    </>
  );
}

function ImportRow({ item, onOpen }: { item: TImport; onOpen: () => void }) {
  const status = statusBadges[item.status];
  return (
    <ListRow
      actions={
        <>
          {item.started_by && (
            <span className="hidden sm:flex">
              <UserChip user={item.started_by} />
            </span>
          )}
          <span className="text-muted-foreground text-xs">
            <RelativeTime value={item.created_at} />
          </span>
          <Button variant="ghost" size="sm" onClick={onOpen}>
            Report
          </Button>
        </>
      }
    >
      <Badge variant={status.variant}>{status.label}</Badge>
      <Badge variant="muted">{item.dry_run ? "Dry run" : "Import"}</Badge>
      <ListRowText primary={reportSummary(item.report)} />
    </ListRow>
  );
}
