import { ListRow, ListRowText } from "@/components/layout/list";
import { GithubIcon } from "@/components/icons/github";
import { RelativeTime } from "@/components/shared/relative-time";
import type { TGithubRepository } from "@/lib/queries/integrations";

export function GithubRepositoryRow({ repository }: { repository: TGithubRepository }) {
  return (
    <ListRow
      actions={
        <span className="text-muted-foreground text-xs">
          <RelativeTime value={repository.last_event_at} />
        </span>
      }
    >
      <GithubIcon className="text-muted-foreground" />
      <ListRowText primary={repository.full_name} secondary={repository.last_event} />
    </ListRow>
  );
}
