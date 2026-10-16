import { readCsrfToken } from "@/lib/server/csrf";
import { getConfig } from "@/lib/config";
import { createClient } from "@/lib/server/client.gen";

const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

export const apiFetch: typeof fetch = async (input, init) => {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers = new Headers(init?.headers);

  if (!SAFE_METHODS.has(method)) {
    const csrf = readCsrfToken();
    if (csrf) headers.set("X-CSRF-Token", csrf);
  }

  return fetch(input, { ...init, credentials: "include", headers });
};

let client: ReturnType<typeof createClient> | null = null;

export function getGoClient() {
  if (!client) {
    client = createClient({ apiUrl: getConfig().apiUrl, fetchFn: apiFetch });
  }
  return client;
}
