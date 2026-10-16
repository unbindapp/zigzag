import { createFileRoute } from "@tanstack/react-router";
import { zodValidator } from "@tanstack/zod-adapter";
import { z } from "zod";

import { INBOX_TABS, InboxView } from "@/components/inbox/inbox-view";

const searchSchema = z.object({
  tab: z.enum(INBOX_TABS).optional(),
  id: z.string().optional(),
});

export const Route = createFileRoute("/$workspace/_workspace/inbox")({
  validateSearch: zodValidator(searchSchema),
  component: InboxPage,
});

function InboxPage() {
  const { tab, id } = Route.useSearch();
  const navigate = Route.useNavigate();
  return (
    <InboxView
      tab={tab ?? "inbox"}
      selectedId={id}
      onTabChange={(next) =>
        void navigate({ search: { tab: next === "inbox" ? undefined : next }, replace: true })
      }
      onSelect={(next) =>
        void navigate({ search: (prev) => ({ ...prev, id: next }), replace: true })
      }
    />
  );
}
