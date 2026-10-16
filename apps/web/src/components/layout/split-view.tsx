import { ReactNode } from "react";

import {
  ResizableGroup,
  ResizableHandle,
  ResizablePanel,
  ResizableSizedPanel,
} from "@/components/ui/resizable";

type TProps = {
  main: ReactNode;
  aside?: ReactNode;
  storageKey?: string;
};

const DEFAULT_ASIDE_WIDTH = 480;
const MIN_ASIDE_WIDTH = 360;
const MAX_ASIDE_WIDTH = 900;

export function SplitView({ main, aside, storageKey }: TProps) {
  return (
    <ResizableGroup
      defaultSize={DEFAULT_ASIDE_WIDTH}
      minSize={MIN_ASIDE_WIDTH}
      maxSize={MAX_ASIDE_WIDTH}
      storageKey={storageKey}
    >
      <ResizablePanel>{main}</ResizablePanel>
      {aside && (
        <>
          <ResizableHandle className="hidden lg:block" />
          <ResizableSizedPanel className="hidden lg:flex">{aside}</ResizableSizedPanel>
        </>
      )}
    </ResizableGroup>
  );
}
