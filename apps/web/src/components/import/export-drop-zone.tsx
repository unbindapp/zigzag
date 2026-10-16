import { FileJsonIcon, UploadIcon } from "lucide-react";
import { DragEvent, useState } from "react";

import { parseLinearExport, type TParsedExport } from "@/components/import/parse-export";
import { Button } from "@/components/ui/button";
import { FileTrigger } from "@/components/ui/file-trigger";
import { cn } from "@/components/ui/utils";

type TProps = {
  selected: TParsedExport | null;
  disabled?: boolean;
  onSelect: (parsed: TParsedExport) => void;
};

export function ExportDropZone({ selected, disabled, onSelect }: TProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const read = async (file: File) => {
    const result = parseLinearExport(file.name, await file.text());
    if (!result.ok) {
      setError(result.error);
      return;
    }
    setError(null);
    onSelect(result.parsed);
  };

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setIsDragging(false);
    const file = event.dataTransfer.files[0];
    if (file && !disabled) void read(file);
  };

  return (
    <div className="flex w-full flex-col gap-2">
      <div
        data-slot="export-drop-zone"
        className={cn(
          "flex w-full flex-col items-center justify-center gap-2 rounded-lg border border-dashed p-4 text-center",
          isDragging && "border-primary bg-accent",
        )}
        onDragOver={(event) => {
          event.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={onDrop}
      >
        <div className="bg-muted text-muted-foreground flex size-10 items-center justify-center rounded-full">
          <FileJsonIcon className="size-5" />
        </div>
        {selected ? (
          <div className="flex min-w-0 flex-col gap-1">
            <p className="truncate text-sm font-medium">{selected.fileName}</p>
            <p className="text-muted-foreground text-xs">{selected.summary}</p>
          </div>
        ) : (
          <div className="flex flex-col gap-1">
            <p className="text-sm font-medium">Drop a Linear export here</p>
            <p className="text-muted-foreground text-sm">
              A JSON file with users, teams and issues.
            </p>
          </div>
        )}
        <FileTrigger accept="application/json,.json" onSelect={(files) => void read(files[0])}>
          {(open) => (
            <Button variant="outline" size="sm" disabled={disabled} onClick={open}>
              <UploadIcon className="size-4" />
              {selected ? "Choose another file" : "Choose file"}
            </Button>
          )}
        </FileTrigger>
      </div>
      {error && (
        <p role="alert" className="text-destructive text-xs">
          {error}
        </p>
      )}
    </div>
  );
}
