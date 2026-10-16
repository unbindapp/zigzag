export const availableThemes = ["dark", "light", "system"] as const;
export const themes = [...availableThemes];
export type TTheme = (typeof availableThemes)[number];
export type TThemeWithoutSystem = Exclude<TTheme, "system">;
export const defaultTheme: TTheme = "system";

export const metaTheme: Record<TThemeWithoutSystem, string> = {
  dark: "#14161B",
  light: "#F9FAFB",
};
