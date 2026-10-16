export const PRESET_COLORS = [
  "#95a2b3",
  "#5e6ad2",
  "#26b5ce",
  "#0f783c",
  "#4cb782",
  "#f2c94c",
  "#f2994a",
  "#eb5757",
  "#db6e9e",
  "#bb87fc",
  "#8e6b4a",
  "#4b5563",
] as const;

export const hexColorPattern = /^#[0-9a-fA-F]{6}$/;

export function isHexColor(value: string | null | undefined): value is string {
  return !!value && hexColorPattern.test(value);
}

function hashString(value: string) {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash * 31 + value.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

export function colorFromId(id: string) {
  return PRESET_COLORS[hashString(id) % PRESET_COLORS.length];
}

export function initials(name: string) {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) return "?";
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}
