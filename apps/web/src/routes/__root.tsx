import type { QueryClient } from "@tanstack/react-query";
import { createRootRouteWithContext, Outlet, redirect } from "@tanstack/react-router";
import { TriangleAlertIcon } from "lucide-react";

import NotFoundTemplate from "@/components/navigation/not-found-template";
import Providers from "@/components/providers/providers";
import { meQuery } from "@/lib/queries/me";

export type RouterContext = {
  queryClient: QueryClient;
};

const PUBLIC_PATHS = new Set(["/sign-in", "/forgot-password", "/reset-password", "/welcome"]);
const PUBLIC_PREFIXES = ["/invite/"];

function isPublicPath(pathname: string) {
  const normalized = pathname.length > 1 ? pathname.replace(/\/$/, "") : pathname;
  return PUBLIC_PATHS.has(normalized) || PUBLIC_PREFIXES.some((p) => pathname.startsWith(p));
}

export const Route = createRootRouteWithContext<RouterContext>()({
  beforeLoad: async ({ context, location }) => {
    if (isPublicPath(location.pathname)) return;
    const me = await context.queryClient.ensureQueryData(meQuery);
    if (!me) {
      throw redirect({ to: "/sign-in", search: { redirect: location.href } });
    }
    return { me };
  },
  component: RootComponent,
  notFoundComponent: () => (
    <main className="flex min-h-svh flex-col">
      <NotFoundTemplate
        code={404}
        description="Not found."
        buttonText="Go home"
        buttonTo="/"
        Icon={TriangleAlertIcon}
      />
    </main>
  ),
  errorComponent: ({ error }) => (
    <main className="flex min-h-svh flex-col">
      <NotFoundTemplate
        code={500}
        description={error instanceof Error ? error.message : "Something went wrong."}
        buttonText="Go home"
        buttonTo="/"
        Icon={TriangleAlertIcon}
      />
    </main>
  ),
});

function RootComponent() {
  return (
    <Providers>
      <Outlet />
    </Providers>
  );
}
