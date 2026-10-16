import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { UserCircleIcon } from "lucide-react";
import { z } from "zod";

import {
  filtersFromSearch,
  filtersToSearch,
  issueSearchSchema,
  type ListIssuesFilters,
  type TFilterField,
} from "@/components/issues/filters";
import { IssuesView } from "@/components/issues/issues-view";
import { Tabs, TabsList, TabsTab } from "@/components/ui/tabs";
import { meQuery } from "@/lib/queries/me";

const TABS = ["assigned", "created", "subscribed"] as const;
type TMyIssuesTab = (typeof TABS)[number];

const tabLabels: Record<TMyIssuesTab, string> = {
  assigned: "Assigned",
  created: "Created",
  subscribed: "Subscribed",
};

const searchSchema = issueSearchSchema.extend({
  tab: z.enum(TABS).optional().catch(undefined),
});

export const Route = createFileRoute("/$workspace/_workspace/my-issues")({
  validateSearch: zodValidator(searchSchema),
  component: MyIssuesPage,
});

function fixedFiltersFor(tab: TMyIssuesTab, userId: string): ListIssuesFilters {
  switch (tab) {
    case "assigned":
      return { assignee_ids: [userId] };
    case "created":
      return { creator_ids: [userId] };
    case "subscribed":
      return { subscriber_id: userId };
  }
}

const hiddenFieldsFor: Record<TMyIssuesTab, TFilterField[]> = {
  assigned: ["assignee"],
  created: ["creator"],
  subscribed: ["subscriber"],
};

function MyIssuesPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: me } = useQuery(meQuery);
  const tab = search.tab ?? "assigned";
  if (!me) return null;
  return (
    <IssuesView
      key={tab}
      title="My issues"
      icon={<UserCircleIcon />}
      storageKey={`my-issues:${tab}`}
      fixedFilters={fixedFiltersFor(tab, me.user.id)}
      hiddenFilterFields={hiddenFieldsFor[tab]}
      defaultDisplay={{ group_by: "state" }}
      filters={filtersFromSearch(search)}
      onFiltersChange={(filters) =>
        void navigate({ search: (prev) => ({ ...prev, ...filtersToSearch(filters) }) })
      }
      tabs={
        <Tabs
          value={tab}
          onValueChange={(value) =>
            void navigate({ search: (prev) => ({ ...prev, tab: value as TMyIssuesTab }) })
          }
        >
          <TabsList>
            {TABS.map((item) => (
              <TabsTab key={item} value={item}>
                {tabLabels[item]}
              </TabsTab>
            ))}
          </TabsList>
        </Tabs>
      }
    />
  );
}
