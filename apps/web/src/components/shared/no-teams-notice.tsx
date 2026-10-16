import { InfoIcon } from "lucide-react";

import { LinkButton } from "@/components/ui/button";
import { useWorkspace } from "@/components/workspace/workspace-provider";

type TProps = {
  entity: "Issues" | "Projects";
  onNavigate?: () => void;
};

export function NoTeamsNotice({ entity, onNavigate }: TProps) {
  const { workspace } = useWorkspace();
  return (
    <div role="status" className="bg-muted flex w-full items-center gap-2 rounded-lg border p-4">
      <InfoIcon className="text-muted-foreground size-4 shrink-0" />
      <p className="min-w-0 flex-1 text-sm">{entity} belong to a team. Create a team first.</p>
      <LinkButton
        to="/$workspace/settings/teams"
        params={{ workspace: workspace.slug }}
        variant="outline"
        size="sm"
        onClick={onNavigate}
      >
        Go to teams
      </LinkButton>
    </div>
  );
}
