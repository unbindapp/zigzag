const layeredSelector = '[role="dialog"], [role="alertdialog"], [role="menu"], [role="listbox"]';

let lastFocusedOutsideLayers: HTMLElement | null = null;

if (typeof document !== "undefined") {
  document.addEventListener(
    "focusin",
    (event) => {
      const target = event.target;
      if (!(target instanceof HTMLElement) || target.closest(layeredSelector)) return;
      lastFocusedOutsideLayers = target;
    },
    true,
  );
}

export function returnFocusTarget(): HTMLElement | true {
  const target = lastFocusedOutsideLayers;
  return target && target.isConnected ? target : true;
}
