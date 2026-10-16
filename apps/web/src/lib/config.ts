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
  try {
    const res = await fetch("/config.json", { cache: "no-cache" });
    if (res.ok) {
      const json = (await res.json()) as Partial<AppConfig>;
      config = { ...fallbackConfig, ...json };
      return config;
    }
  } catch {
    // fall through to the build-time fallback
  }
  config = fallbackConfig;
  return config;
}

export function getConfig(): AppConfig {
  if (!config) {
    throw new Error("Config not loaded yet. loadConfig() must run before getConfig().");
  }
  return config;
}
