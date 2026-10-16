import * as React from "react";
import { useLocalStorage } from "usehooks-ts";

import { cn } from "@/components/ui/utils";

type TResizableContext = {
  size: number;
  setSize: (size: number) => void;
  minSize: number;
  maxSize: number;
};

const ResizableContext = React.createContext<TResizableContext | null>(null);

function useResizable() {
  const context = React.useContext(ResizableContext);
  if (!context) throw new Error("useResizable must be used within ResizableGroup");
  return context;
}

const KEYBOARD_STEP = 24;

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function ResizableGroup({
  defaultSize,
  minSize,
  maxSize,
  storageKey,
  className,
  children,
}: {
  defaultSize: number;
  minSize: number;
  maxSize: number;
  storageKey?: string;
  className?: string;
  children: React.ReactNode;
}) {
  const [storedSize, setStoredSize] = useLocalStorage(
    storageKey ?? "resizable:default",
    defaultSize,
  );
  const [localSize, setLocalSize] = React.useState(defaultSize);
  const size = clamp(storageKey ? storedSize : localSize, minSize, maxSize);
  const setSize = storageKey ? setStoredSize : setLocalSize;

  const value = React.useMemo(
    () => ({ size, setSize: (s: number) => setSize(clamp(s, minSize, maxSize)), minSize, maxSize }),
    [size, setSize, minSize, maxSize],
  );

  return (
    <ResizableContext.Provider value={value}>
      <div data-slot="resizable-group" className={cn("flex min-h-0 w-full flex-1", className)}>
        {children}
      </div>
    </ResizableContext.Provider>
  );
}

function ResizablePanel({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="resizable-panel"
      className={cn("flex min-h-0 min-w-0 flex-1 flex-col", className)}
      {...props}
    />
  );
}

function ResizableSizedPanel({ className, style, ...props }: React.ComponentProps<"div">) {
  const { size } = useResizable();
  return (
    <div
      data-slot="resizable-sized-panel"
      className={cn("flex min-h-0 shrink-0 flex-col", className)}
      style={{ ...style, width: size }}
      {...props}
    />
  );
}

function ResizableHandle({
  className,
  label = "Resize panel",
}: {
  className?: string;
  label?: string;
}) {
  const { size, setSize, minSize, maxSize } = useResizable();
  const startRef = React.useRef<{ x: number; size: number } | null>(null);

  const onPointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId);
    startRef.current = { x: event.clientX, size };
  };

  const onPointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    const start = startRef.current;
    if (!start) return;
    setSize(start.size + (start.x - event.clientX));
  };

  const onPointerUp = (event: React.PointerEvent<HTMLDivElement>) => {
    event.currentTarget.releasePointerCapture(event.pointerId);
    startRef.current = null;
  };

  const onKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    switch (event.key) {
      case "ArrowLeft":
        event.preventDefault();
        setSize(size + KEYBOARD_STEP);
        return;
      case "ArrowRight":
        event.preventDefault();
        setSize(size - KEYBOARD_STEP);
        return;
    }
  };

  return (
    <div
      role="separator"
      aria-orientation="vertical"
      aria-label={label}
      aria-valuenow={size}
      aria-valuemin={minSize}
      aria-valuemax={maxSize}
      tabIndex={0}
      onPointerDown={onPointerDown}
      onPointerMove={onPointerMove}
      onPointerUp={onPointerUp}
      onKeyDown={onKeyDown}
      data-slot="resizable-handle"
      className={cn(
        "bg-border has-hover:hover:bg-primary/8-10 focus-visible:bg-primary/8-10 relative w-px shrink-0 cursor-col-resize touch-none outline-hidden transition-colors before:absolute before:inset-y-0 before:-left-1 before:w-2",
        className,
      )}
    />
  );
}

export { ResizableGroup, ResizablePanel, ResizableSizedPanel, ResizableHandle };
