export const PULL_REQUEST_STATES = ["open", "draft", "merged", "closed"] as const;
export type TPullRequestState = (typeof PULL_REQUEST_STATES)[number];

export type TGithubLink =
  | { kind: "pull_request"; reference: string; state: TPullRequestState | null }
  | { kind: "branch"; reference: string }
  | { kind: "commit"; reference: string; sha: string; subject: string };

type TLinkedAttachment = { source_type: string; title: string; metadata: Record<string, unknown> };

const SHORT_SHA_LENGTH = 7;

function text(value: unknown) {
  return typeof value === "string" ? value : "";
}

export function githubLinkOf(attachment: TLinkedAttachment): TGithubLink | null {
  const repo = text(attachment.metadata.repo);
  switch (attachment.source_type) {
    case "github": {
      const number = attachment.metadata.number;
      const state = PULL_REQUEST_STATES.find((value) => value === attachment.metadata.state);
      return {
        kind: "pull_request",
        reference: typeof number === "number" ? `${repo}#${number}` : repo,
        state: state ?? null,
      };
    }
    case "github_branch":
      return { kind: "branch", reference: repo };
    case "github_commit": {
      const sha = text(attachment.metadata.sha).slice(0, SHORT_SHA_LENGTH);
      const subject = attachment.title.startsWith(`${sha} `)
        ? attachment.title.slice(sha.length + 1)
        : attachment.title;
      return { kind: "commit", reference: repo, sha, subject };
    }
    default:
      return null;
  }
}
