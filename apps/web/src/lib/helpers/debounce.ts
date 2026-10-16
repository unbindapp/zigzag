export type TDebounced<Args extends unknown[]> = {
  call: (...args: Args) => void;
  flush: () => void;
  cancel: () => void;
  isPending: () => boolean;
};

export function createDebounce<Args extends unknown[]>(
  fn: (...args: Args) => void,
  delayMs: number,
): TDebounced<Args> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  let pendingArgs: Args | undefined;

  const cancel = () => {
    clearTimeout(timer);
    timer = undefined;
    pendingArgs = undefined;
  };

  const flush = () => {
    const args = pendingArgs;
    cancel();
    if (args) fn(...args);
  };

  const call = (...args: Args) => {
    clearTimeout(timer);
    pendingArgs = args;
    timer = setTimeout(flush, delayMs);
  };

  return { call, flush, cancel, isPending: () => pendingArgs !== undefined };
}
