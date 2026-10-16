import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { ListTodoIcon } from "lucide-react";

import { filtersFromSearch, filtersToSearch, issueSearchSchema } from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";
import type { TStateType } from "@/components/shared/state-icon";
import { teamCrumb } from "@/components/workspace/team-crumb";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";

const STATE_TYPES: TStateType[] = ["backlog"];

export const Route = createFileRoute("/$workspace/_workspace/team/$key/backlog")({
  validateSearch: zodValidator(issueSearchSchema),
  component: TeamKeyBacklogPage,
});

function TeamKeyBacklogPage() {
  const team = useRequiredTeamFromRoute();
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { workspace } = Route.useParams();
  return (
    <IssuesView
      key={team.id}
      title="Backlog"
      icon={<ListTodoIcon />}
      breadcrumb={[teamCrumb(team, workspace), { label: "Backlog" }]}
      storageKey={`team:${team.id}:backlog`}
      teamId={team.id}
      fixedFilters={{ team_ids: [team.id], state_types: STATE_TYPES }}
      filters={filtersFromSearch(search)}
      onFiltersChange={(filters) =>
        void navigate({ search: (prev) => ({ ...prev, ...filtersToSearch(filters) }) })
      }
    />
  );
}
