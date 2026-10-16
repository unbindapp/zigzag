export type THotkeyScope = { id: string; keys: string[]; modal: boolean };

export function normalizeHotkey(hotkey: string) {
  return hotkey
    .split("+")
    .map((part) => part.trim().toLowerCase())
    .sort()
    .join("+");
}

export function hotkeyOwners(scopes: THotkeyScope[]) {
  const lastModal = scopes.findLastIndex((scope) => scope.modal);
  const owners = new Map<string, string>();
  for (const scope of scopes.slice(Math.max(lastModal, 0))) {
    for (const key of scope.keys) owners.set(normalizeHotkey(key), scope.id);
  }
  return owners;
}
