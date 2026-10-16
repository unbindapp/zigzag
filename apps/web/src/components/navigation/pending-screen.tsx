import { LoaderIcon } from "lucide-react";

export default function PendingScreen() {
  return (
    <div
      aria-busy
      aria-label="Loading"
      className="text-muted-foreground flex w-full flex-1 items-center justify-center"
    >
      <LoaderIcon className="size-5 animate-spin" />
    </div>
  );
}
