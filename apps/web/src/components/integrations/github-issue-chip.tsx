import { GithubIcon } from "@/components/icons/github";
import { badgeVariants } from "@/components/ui/badge";
import { SimpleTooltip } from "@/components/ui/tooltip";
import type { TIssueDetail } from "@/lib/queries/issues";

type TGithubIssueLink = NonNullable<TIssueDetail["github_issue"]>;

export function GithubIssueChip({ link }: { link: TGithubIssueLink }) {
  return (
    <SimpleTooltip content={`Synced with ${link.repository}#${link.number}`}>
      <a
        href={link.html_url}
        target="_blank"
        rel="noreferrer noopener"
        className={badgeVariants({ variant: "outline" })}
      >
        <GithubIcon />
        <span className="font-mono">GitHub #{link.number}</span>
      </a>
    </SimpleTooltip>
  );
}
