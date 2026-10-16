export const TEAM_KEY_MAX_LENGTH = 7;
export const teamKeyPattern = /^[A-Z][A-Z0-9]{0,6}$/;
export const workspaceSlugPattern = /^[a-z0-9-]{2,32}$/;

const identifierPattern = /^([A-Za-z][A-Za-z0-9]{0,6})-(\d+)$/;

export function parseIdentifier(value: string): { key: string; number: number } | null {
  const match = value.trim().match(identifierPattern);
  if (!match) return null;
  return { key: match[1].toUpperCase(), number: Number(match[2]) };
}

export function formatIdentifier(key: string, number: number) {
  return `${key.toUpperCase()}-${number}`;
}

export function deriveTeamKey(name: string) {
  const words = name
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^A-Za-z0-9\s]/g, " ")
    .split(/\s+/)
    .filter(Boolean);
  const firstLetterWord = words.findIndex((word) => /^[A-Za-z]/.test(word));
  if (firstLetterWord === -1) return "";
  const usable = words.slice(firstLetterWord);
  const raw = usable.length === 1 ? usable[0].slice(0, 3) : usable.map((word) => word[0]).join("");
  return raw.toUpperCase().slice(0, TEAM_KEY_MAX_LENGTH);
}

export function sanitizeTeamKey(value: string) {
  return value
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "")
    .replace(/^[0-9]+/, "")
    .slice(0, TEAM_KEY_MAX_LENGTH);
}

export function slugify(value: string, maxLength = 32) {
  return value
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, maxLength)
    .replace(/-+$/g, "");
}
