import { z } from "zod";

export const githubLoginSchema = z
  .string()
  .trim()
  .transform((value) => value.replace(/^@/, "").toLowerCase())
  .refine((value) => value === "" || /^[a-z0-9](?:[a-z0-9]|-(?=[a-z0-9])){0,38}$/.test(value), {
    message: "Use the GitHub username: letters, digits and single hyphens.",
  });

export const repositoryNameSchema = z
  .string()
  .trim()
  .transform((value) =>
    value
      .replace(/^https:\/\/github\.com\//i, "")
      .replace(/\.git$/i, "")
      .replace(/\/+$/, ""),
  )
  .refine((value) => /^[\w.-]+\/[\w.-]+$/.test(value), {
    message: "Enter the repository as owner/name.",
  });

export const githubTokenSchema = z
  .string()
  .trim()
  .min(20, "Paste the whole token.")
  .refine((value) => !/\s/.test(value), { message: "The token has no spaces." });

type TBackfill = {
  backfill_status: "none" | "running" | "completed" | "failed";
  backfill_report?: {
    issues_created: number;
    issues_updated: number;
    comments_created: number;
    skipped: number;
  };
};

export function isImporting(repoSync: TBackfill) {
  return repoSync.backfill_status === "running";
}

export function importSummary(repoSync: TBackfill): string | null {
  const report = repoSync.backfill_report;
  if (!report || repoSync.backfill_status === "none") return null;
  const parts = [
    `${report.issues_created} created`,
    `${report.issues_updated} updated`,
    `${report.comments_created} ${report.comments_created === 1 ? "comment" : "comments"}`,
    `${report.skipped} skipped`,
  ];
  return parts.join(", ");
}

export function linkedIssuesLabel(count: number) {
  return count === 1 ? "1 issue" : `${count} issues`;
}

const githubCommentPrefix = /^\*\*@([a-z0-9-]+) on GitHub\*\*\n+/;

export function githubCommentAuthor(body: string): { login: string; body: string } | null {
  const match = githubCommentPrefix.exec(body);
  if (!match) return null;
  return { login: match[1], body: body.slice(match[0].length) };
}
