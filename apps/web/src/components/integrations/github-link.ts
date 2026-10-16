export const PULL_REQUEST_STATES = ["open", "draft", "merged", "closed"] as const;
export type TPullRequestState = (typeof PULL_REQUEST_STATES)[number];

export type TGithubLink =
  | { kind: "pull_request"; reference: string; state: TPullRequestState | null }
  | { kind: "branch"; reference: string };

type TLinkedAttachment = { source_type: string; metadata: Record<string, unknown> };

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
    default:
      return null;
  }
}
