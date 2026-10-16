export type AppConfig = {
  apiUrl: string;
};

const fallbackConfig: AppConfig = {
  apiUrl: import.meta.env.VITE_API_URL ?? "/api/go",
};

let config: AppConfig | null = null;

export async function loadConfig(): Promise<AppConfig> {
  if (import.meta.env.DEV) {
    config = { apiUrl: "/api/go" };
    return config;
  }
  const served = await fetchServedConfig().catch(() => null);
  config = { ...fallbackConfig, ...served };
  return config;
}

async function fetchServedConfig(): Promise<Partial<AppConfig> | null> {
  const res = await fetch("/config.json", { cache: "no-cache" });
  if (!res.ok) return null;
  return (await res.json()) as Partial<AppConfig>;
}

export function getConfig(): AppConfig {
  if (!config) {
    throw new Error("Config not loaded yet. loadConfig() must run before getConfig().");
  }
  return config;
}
