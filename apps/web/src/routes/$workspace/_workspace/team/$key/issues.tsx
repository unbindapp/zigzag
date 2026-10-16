import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { CircleDotIcon } from "lucide-react";

import { filtersFromSearch, filtersToSearch, issueSearchSchema } from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";
import type { TStateType } from "@/components/shared/state-icon";
import { TeamIcon } from "@/components/shared/team-icon";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";

const STATE_TYPES: TStateType[] = ["unstarted", "started", "completed", "canceled"];

export const Route = createFileRoute("/$workspace/_workspace/team/$key/issues")({
  validateSearch: zodValidator(issueSearchSchema),
  component: TeamKeyIssuesPage,
});

function TeamKeyIssuesPage() {
  const team = useRequiredTeamFromRoute();
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { workspace } = Route.useParams();
  return (
    <IssuesView
      key={team.id}
      title="Issues"
      icon={<CircleDotIcon />}
      breadcrumb={[
        {
          label: team.name,
          icon: <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />,
          link: { to: "/$workspace/team/$key/issues", params: { workspace, key: team.key } },
        },
        { label: "Issues" },
      ]}
      storageKey={`team:${team.id}:issues`}
      teamId={team.id}
      fixedFilters={{ team_ids: [team.id], state_types: STATE_TYPES }}
      filters={filtersFromSearch(search)}
      onFiltersChange={(filters) =>
        void navigate({ search: (prev) => ({ ...prev, ...filtersToSearch(filters) }) })
      }
    />
  );
}
