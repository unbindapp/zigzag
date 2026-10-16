import { MutationCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { createRouter, RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import "@/globals.css";
import PendingScreen from "@/components/navigation/pending-screen";
import { toast } from "@/components/ui/toast";
import { loadConfig } from "@/lib/config";
import { errorText, isApiError } from "@/lib/server/errors";
import { routeTree } from "./routeTree.gen";

const DEFAULT_STALE_TIME_MS = 5 * 1000;
const DEFAULT_PENDING_MS = 1000;
const DEFAULT_PENDING_MIN_MS = 250;
const DEFAULT_INTENT_DELAY_MS = 50;

declare module "@tanstack/react-query" {
  interface Register {
    mutationMeta: {
      errorTitle?: string;
      silent?: boolean;
    };
  }
}

const MAX_QUERY_RETRIES = 3;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: DEFAULT_STALE_TIME_MS,
      retry: (failureCount, error) =>
        !(isApiError(error) && error.status < 500) && failureCount < MAX_QUERY_RETRIES,
    },
  },
  mutationCache: new MutationCache({
    onError: (error, _variables, _onMutateResult, mutation) => {
      if (mutation.meta?.silent) return;
      toast.add({
        type: "error",
        title: mutation.meta?.errorTitle ?? "Something went wrong",
        description: errorText(error),
      });
    },
  }),
});

const router = createRouter({
  routeTree,
  context: { queryClient },
  defaultPreload: "intent",
  defaultPreloadStaleTime: DEFAULT_STALE_TIME_MS,
  defaultPreloadDelay: DEFAULT_INTENT_DELAY_MS,
  defaultPendingComponent: PendingScreen,
  defaultPendingMs: DEFAULT_PENDING_MS,
  defaultPendingMinMs: DEFAULT_PENDING_MIN_MS,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

async function bootstrap() {
  await loadConfig();

  createRoot(document.getElementById("root")!).render(
    <StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </StrictMode>,
  );
}

void bootstrap();
