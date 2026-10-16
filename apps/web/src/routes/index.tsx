import { createFileRoute, redirect } from "@tanstack/react-router";

import { getLastWorkspaceSlug } from "@/lib/last-workspace";
import { meQuery, meWorkspaces } from "@/lib/queries/me";

export const Route = createFileRoute("/")({
  beforeLoad: async ({ context }) => {
    const me = await context.queryClient.ensureQueryData(meQuery);
    const workspaces = meWorkspaces(me);
    const lastSlug = getLastWorkspaceSlug();
    const workspace = workspaces.find((w) => w.slug === lastSlug) ?? workspaces[0];
    if (!workspace) throw redirect({ to: "/create-workspace" });
    throw redirect({ to: "/$workspace", params: { workspace: workspace.slug } });
  },
});
