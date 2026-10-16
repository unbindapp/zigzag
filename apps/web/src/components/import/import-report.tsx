import { List, ListRow, ListRowText } from "@/components/layout/list";
import { SectionTitle } from "@/components/layout/section-title";
import { countRows } from "@/components/import/parse-export";
import { Badge } from "@/components/ui/badge";
import { SimpleTooltip } from "@/components/ui/tooltip";
import type { TImportReport } from "@/lib/queries/imports";

type TMappingAction = TImportReport["user_mapping"][number]["action"];

const mappingBadges: Record<
  TMappingAction,
  { label: string; variant: "success" | "process" | "muted" }
> = {
  matched: { label: "Matched", variant: "success" },
  created: { label: "Placeholder", variant: "process" },
  skipped: { label: "Skipped", variant: "muted" },
};

function CountCell({ value }: { value: number }) {
  return (
    <span
      className={
        value === 0 ? "text-muted-foreground w-16 text-right text-xs" : "w-16 text-right text-xs"
      }
    >
      {value}
    </span>
  );
}

export function ImportCounts({ report }: { report: Pick<TImportReport, "counts"> }) {
  return (
    <List>
      <ListRow
        muted
        actions={
          <>
            <span className="w-16 text-right text-xs">Created</span>
            <span className="w-16 text-right text-xs">Updated</span>
            <span className="w-16 text-right text-xs">Skipped</span>
          </>
        }
      >
        <span className="text-xs">Entity</span>
      </ListRow>
      {countRows(report).map((row) => (
        <ListRow
          key={row.key}
          actions={
            <>
              <CountCell value={row.created} />
              <CountCell value={row.updated} />
              <CountCell value={row.skipped} />
            </>
          }
        >
          <ListRowText primary={row.label} />
        </ListRow>
      ))}
    </List>
  );
}

export function ImportWarnings({ warnings }: { warnings: TImportReport["warnings"] }) {
  if (warnings.length === 0) {
    return <p className="text-muted-foreground text-sm">No warnings.</p>;
  }
  return (
    <List>
      {warnings.map((warning, index) => (
        <ListRow
          key={`${warning.entity}-${warning.external_id}-${index}`}
          actions={
            <Badge variant="warning" className="hidden sm:inline-flex">
              {warning.entity}
            </Badge>
          }
        >
          {warning.external_id && (
            <span className="text-muted-foreground w-16 shrink-0 truncate font-mono text-xs">
              {warning.external_id}
            </span>
          )}
          <SimpleTooltip content={warning.message}>
            <span className="min-w-0 flex-1 truncate">{warning.message}</span>
          </SimpleTooltip>
        </ListRow>
      ))}
    </List>
  );
}

export function UserMapping({ mapping }: { mapping: TImportReport["user_mapping"] }) {
  if (mapping.length === 0) {
    return <p className="text-muted-foreground text-sm">The export lists no users.</p>;
  }
  return (
    <List>
      {mapping.map((entry) => {
        const badge = mappingBadges[entry.action];
        return (
          <ListRow
            key={entry.linear_id}
            actions={<Badge variant={badge.variant}>{badge.label}</Badge>}
          >
            <ListRowText primary={entry.email} secondary={entry.linear_id} />
          </ListRow>
        );
      })}
    </List>
  );
}

export function ImportReportView({ report }: { report: TImportReport }) {
  return (
    <div className="flex w-full flex-col gap-6">
      <div className="flex w-full flex-col gap-4">
        <SectionTitle
          title="Counts"
          description={
            report.dry_run
              ? "What a real import would do. Nothing was written."
              : "What the import created, updated and skipped."
          }
          actions={
            <Badge variant={report.dry_run ? "muted" : "success"}>
              {report.dry_run ? "Dry run" : "Imported"}
            </Badge>
          }
        />
        <ImportCounts report={report} />
      </div>
      <div className="flex w-full flex-col gap-4">
        <SectionTitle
          title="Warnings"
          description="Rows the import changed or left out, and why."
          actions={<Badge variant="muted">{report.warnings.length}</Badge>}
        />
        <ImportWarnings warnings={report.warnings} />
      </div>
      <div className="flex w-full flex-col gap-4">
        <SectionTitle
          title="People"
          description="How each Linear user maps to an account in this instance."
        />
        <UserMapping mapping={report.user_mapping} />
      </div>
    </div>
  );
}
