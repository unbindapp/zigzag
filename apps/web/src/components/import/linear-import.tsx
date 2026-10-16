import { DatabaseIcon, FlaskConicalIcon } from "lucide-react";
import { useState } from "react";

import { ExportDropZone } from "@/components/import/export-drop-zone";
import { ImportHistory } from "@/components/import/import-history";
import {
  DEFAULT_IMPORT_OPTIONS,
  ImportOptions,
  type TImportOptions,
} from "@/components/import/import-options";
import { ImportReportView } from "@/components/import/import-report";
import { reportSummary, type TParsedExport } from "@/components/import/parse-export";
import { SettingsSection } from "@/components/layout/settings-section";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { type TImportReport, useImportLinearMutation } from "@/lib/queries/imports";

type TRun = {
  file: TParsedExport;
  options: TImportOptions;
  report: TImportReport;
};

export function LinearImport() {
  const { workspace } = useWorkspace();
  const [file, setFile] = useState<TParsedExport | null>(null);
  const [options, setOptions] = useState<TImportOptions>(DEFAULT_IMPORT_OPTIONS);
  const [run, setRun] = useState<TRun | null>(null);
  const { mutate, isPending, variables } = useImportLinearMutation();

  const dryRunMatches =
    run !== null && run.report.dry_run && run.file === file && run.options === options;

  const start = (dryRun: boolean) => {
    if (!file) return;
    mutate(
      {
        workspace_id: workspace.id,
        dry_run: dryRun,
        restore_archived: options.restoreArchived,
        restore_trashed: options.restoreTrashed,
        create_missing_users: options.createMissingUsers,
        export: file.export,
      },
      {
        onSuccess: ({ data }) => {
          setRun({ file, options, report: data });
          if (!dryRun) {
            toast.add({
              type: "success",
              title: "Linear export imported",
              description: reportSummary(data),
            });
          }
        },
      },
    );
  };

  return (
    <>
      <SettingsSection
        title="Linear export"
        description="Upload the JSON export of a Linear workspace. Teams match by key, people by email, and issues keep their numbers."
      >
        <ExportDropZone
          selected={file}
          disabled={isPending}
          onSelect={(parsed) => {
            setFile(parsed);
            setRun(null);
          }}
        />
      </SettingsSection>
      <SettingsSection
        title="Options"
        description="Run a dry run first: it does the whole import and rolls it back, so its report is exact."
        footer={
          <>
            <Button
              variant={dryRunMatches ? "outline" : "default"}
              size="sm"
              disabled={!file}
              isPending={isPending && variables?.dry_run === true}
              onClick={() => start(true)}
            >
              <FlaskConicalIcon className="size-4" />
              Dry run
            </Button>
            {dryRunMatches && (
              <Button
                size="sm"
                isPending={isPending && variables?.dry_run === false}
                disabled={isPending}
                onClick={() => start(false)}
              >
                <DatabaseIcon className="size-4" />
                Import for real
              </Button>
            )}
          </>
        }
      >
        <ImportOptions value={options} disabled={isPending} onChange={setOptions} />
      </SettingsSection>
      {run && (
        <SettingsSection
          title={run.report.dry_run ? "Dry run report" : "Import report"}
          description={`${run.file.fileName}: ${reportSummary(run.report)}`}
        >
          <ImportReportView report={run.report} />
        </SettingsSection>
      )}
      <SettingsSection
        title="History"
        description="The latest imports of this workspace, newest first."
      >
        <ImportHistory />
      </SettingsSection>
    </>
  );
}
