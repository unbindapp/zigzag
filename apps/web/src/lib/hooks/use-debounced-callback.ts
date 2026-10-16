import { useEffect, useRef, useState } from "react";

import { createDebounce } from "@/lib/helpers/debounce";

export default function useDebouncedCallback<Args extends unknown[]>(
  fn: (...args: Args) => void,
  delayMs: number,
) {
  const fnRef = useRef(fn);
  useEffect(() => {
    fnRef.current = fn;
  });
  const [debounced] = useState(() =>
    createDebounce((...args: Args) => fnRef.current(...args), delayMs),
  );
  useEffect(() => () => debounced.flush(), [debounced]);
  return debounced;
}
