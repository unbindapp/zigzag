import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { TriangleIcon } from "lucide-react";

import { filtersFromSearch, filtersToSearch, issueSearchSchema } from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";
import { EmptyState } from "@/components/layout/empty-state";
import { PageContainer } from "@/components/layout/page-container";
import { PageHeader } from "@/components/layout/page-header";
import { LinkButton } from "@/components/ui/button";
import type { TStateType } from "@/components/shared/state-icon";
import { teamCrumb } from "@/components/workspace/team-crumb";
import { useRequiredTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useWorkspace } from "@/components/workspace/workspace-provider";

const STATE_TYPES: TStateType[] = ["triage"];

export const Route = createFileRoute("/$workspace/_workspace/team/$key/triage")({
  validateSearch: zodValidator(issueSearchSchema),
  component: TeamKeyTriagePage,
});

function TeamKeyTriagePage() {
  const team = useRequiredTeamFromRoute();
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { workspace } = Route.useParams();
  const { isAdmin } = useWorkspace();
  if (!team.triage_enabled) {
    return (
      <PageContainer>
        <PageHeader
          title="Triage"
          icon={<TriangleIcon />}
          breadcrumb={[teamCrumb(team, workspace), { label: "Triage" }]}
        />
        <EmptyState
          icon={TriangleIcon}
          title="Triage is off for this team"
          description="With triage on, issues from outside the team land here first."
          action={
            isAdmin && (
              <LinkButton
                to="/$workspace/settings/team/$key/triage"
                params={{ workspace, key: team.key }}
                variant="outline"
                size="sm"
              >
                Turn on triage
              </LinkButton>
            )
          }
        />
      </PageContainer>
    );
  }
  return (
    <IssuesView
      key={team.id}
      title="Triage"
      icon={<TriangleIcon />}
      breadcrumb={[teamCrumb(team, workspace), { label: "Triage" }]}
      storageKey={`team:${team.id}:triage`}
      teamId={team.id}
      fixedFilters={{ team_ids: [team.id], state_types: STATE_TYPES }}
      filters={filtersFromSearch(search)}
      onFiltersChange={(filters) =>
        void navigate({ search: (prev) => ({ ...prev, ...filtersToSearch(filters) }) })
      }
    />
  );
}
