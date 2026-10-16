import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

import { TeamsProvider } from "@/components/workspace/teams-provider";
import { WorkspaceProvider } from "@/components/workspace/workspace-provider";
import { WorkspaceShell } from "@/components/workspace/workspace-shell";
import { useRememberWorkspace } from "@/lib/last-workspace";
import { isSystemAdmin, meQuery, meWorkspaces } from "@/lib/queries/me";

export const Route = createFileRoute("/$workspace/_workspace")({
  beforeLoad: async ({ context, params }) => {
    const cached = await context.queryClient.ensureQueryData(meQuery);
    const findWorkspace = (me: typeof cached) =>
      meWorkspaces(me).find((w) => w.slug === params.workspace);
    const me = findWorkspace(cached) ? cached : await context.queryClient.fetchQuery(meQuery);
    if (!me) throw redirect({ to: "/sign-in" });
    const workspace = findWorkspace(me);
    if (workspace) return { workspace };
    const first = meWorkspaces(me)[0];
    if (!first) throw redirect({ to: "/create-workspace" });
    throw redirect({ to: "/$workspace", params: { workspace: first.slug } });
  },
  component: WorkspaceLayout,
});

function WorkspaceLayout() {
  const { workspace: initial } = Route.useRouteContext();
  const { data: me } = useQuery(meQuery);
  const workspace = meWorkspaces(me).find((w) => w.id === initial.id) ?? initial;
  useRememberWorkspace(workspace.slug);

  return (
    <WorkspaceProvider workspace={workspace} isSystemAdmin={isSystemAdmin(me)}>
      <TeamsProvider>
        <WorkspaceShell>
          <Outlet />
        </WorkspaceShell>
      </TeamsProvider>
    </WorkspaceProvider>
  );
}
