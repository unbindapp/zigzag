import { type HotkeyCallback, useHotkeys } from "@tanstack/react-hotkeys";
import { useEffect, useId } from "react";
import { create } from "zustand";

import { hotkeyOwners, normalizeHotkey, type THotkeyScope } from "@/lib/hotkeys/scopes";

type TStore = {
  scopes: THotkeyScope[];
  activate: (scope: THotkeyScope) => void;
  deactivate: (id: string) => void;
};

const useHotkeyScopeStore = create<TStore>((set) => ({
  scopes: [],
  activate: (scope) =>
    set((state) => ({ scopes: [...state.scopes.filter((item) => item.id !== scope.id), scope] })),
  deactivate: (id) => set((state) => ({ scopes: state.scopes.filter((item) => item.id !== id) })),
}));

export type THotkeyBinding = {
  hotkey: string;
  callback: HotkeyCallback;
  enabled?: boolean;
  ignoreInputs?: boolean;
};

type TOptions = { enabled?: boolean; modal?: boolean };

export function useHotkeyScope(
  bindings: THotkeyBinding[],
  { enabled = true, modal = false }: TOptions = {},
) {
  const id = useId();
  const activate = useHotkeyScopeStore((state) => state.activate);
  const deactivate = useHotkeyScopeStore((state) => state.deactivate);
  const keys = bindings
    .filter((binding) => binding.enabled !== false)
    .map((binding) => binding.hotkey);
  const keysSignature = keys.join(" ");

  useEffect(() => {
    if (!enabled) return;
    activate({ id, keys: keysSignature ? keysSignature.split(" ") : [], modal });
    return () => deactivate(id);
  }, [id, enabled, modal, keysSignature, activate, deactivate]);

  const ownedSignature = useHotkeyScopeStore((state) =>
    [...hotkeyOwners(state.scopes)]
      .filter(([, owner]) => owner === id)
      .map(([key]) => key)
      .join(" "),
  );
  const owned = new Set(ownedSignature.split(" "));

  useHotkeys(
    bindings.map((binding) => ({
      hotkey: binding.hotkey as Parameters<typeof useHotkeys>[0][number]["hotkey"],
      callback: binding.callback,
      options: {
        enabled: enabled && binding.enabled !== false && owned.has(normalizeHotkey(binding.hotkey)),
        ...(binding.ignoreInputs === undefined ? {} : { ignoreInputs: binding.ignoreInputs }),
      },
    })),
    { conflictBehavior: "allow" },
  );
}
