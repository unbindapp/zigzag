import type { TImportReport, TLinearExport } from "@/lib/queries/imports";

export type TParsedExport = {
  fileName: string;
  export: TLinearExport;
  summary: string;
};

export type TParseResult = { ok: true; parsed: TParsedExport } | { ok: false; error: string };

function plural(count: number, one: string, many: string) {
  return `${count} ${count === 1 ? one : many}`;
}

function listLength(value: unknown) {
  return Array.isArray(value) ? value.length : 0;
}

export function parseLinearExport(fileName: string, text: string): TParseResult {
  let data: unknown;
  try {
    data = JSON.parse(text);
  } catch {
    return { ok: false, error: `${fileName} is not a JSON file.` };
  }
  if (typeof data !== "object" || data === null || Array.isArray(data)) {
    return { ok: false, error: `${fileName} is not a Linear export.` };
  }
  const record = data as Record<string, unknown>;
  if (!Array.isArray(record.issues) || !Array.isArray(record.teams)) {
    return {
      ok: false,
      error: `${fileName} has no teams and issues lists, so it is not a Linear export.`,
    };
  }
  const workspace = record.workspace as { name?: unknown } | undefined;
  const source =
    typeof workspace?.name === "string" && workspace.name ? ` from ${workspace.name}` : "";
  const summary = [
    plural(listLength(record.issues), "issue", "issues"),
    plural(listLength(record.teams), "team", "teams"),
    plural(listLength(record.projects), "project", "projects"),
    plural(listLength(record.users), "user", "users"),
  ].join(", ");
  return {
    ok: true,
    parsed: { fileName, export: data as TLinearExport, summary: `${summary}${source}` },
  };
}

export const IMPORT_ENTITIES = [
  { key: "users", label: "Users" },
  { key: "teams", label: "Teams" },
  { key: "workflow_states", label: "Workflow states" },
  { key: "labels", label: "Labels" },
  { key: "projects", label: "Projects" },
  { key: "milestones", label: "Milestones" },
  { key: "cycles", label: "Cycles" },
  { key: "issues", label: "Issues" },
  { key: "comments", label: "Comments" },
  { key: "attachments", label: "Attachments" },
  { key: "relations", label: "Relations" },
  { key: "history", label: "History entries" },
] as const;

export function countRows(report: Pick<TImportReport, "counts">) {
  return IMPORT_ENTITIES.map(({ key, label }) => {
    const count = report.counts[key] ?? { created: 0, updated: 0, skipped: 0 };
    return { key, label, ...count };
  });
}

export function reportSummary(report: Pick<TImportReport, "counts">) {
  const issues = report.counts.issues ?? { created: 0, updated: 0, skipped: 0 };
  const parts = [];
  if (issues.created > 0) parts.push(`${issues.created} created`);
  if (issues.updated > 0) parts.push(`${issues.updated} updated`);
  if (issues.skipped > 0) parts.push(`${issues.skipped} skipped`);
  if (parts.length === 0) return "No issues";
  return `Issues: ${parts.join(", ")}`;
}
