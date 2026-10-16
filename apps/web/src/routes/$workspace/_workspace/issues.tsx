import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { ListTodoIcon } from "lucide-react";

import { filtersFromSearch, filtersToSearch, issueSearchSchema } from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";

export const Route = createFileRoute("/$workspace/_workspace/issues")({
  validateSearch: zodValidator(issueSearchSchema),
  component: IssuesPage,
});

function IssuesPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  return (
    <IssuesView
      title="All issues"
      icon={<ListTodoIcon />}
      storageKey="all-issues"
      filters={filtersFromSearch(search)}
      onFiltersChange={(filters) =>
        void navigate({ search: (prev) => ({ ...prev, ...filtersToSearch(filters) }) })
      }
    />
  );
}
