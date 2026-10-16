import { useState } from "react";

import { HEALTHS, HealthIcon, healthLabels } from "@/components/projects/health-badge";
import { MarkdownEditor } from "@/components/shared/markdown-editor";
import { type TUserSummary } from "@/components/shared/user-chip";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { THealth } from "@/lib/queries/projects";

type TProps = {
  initialBody?: string;
  initialHealth?: THealth;
  submitLabel?: string;
  users?: TUserSummary[];
  onSubmit: (value: { body: string; health: THealth }) => Promise<unknown>;
  onCancel?: () => void;
};

const healthItems = HEALTHS.map((health) => ({ value: health, label: healthLabels[health] }));

export function StatusUpdateComposer({
  initialBody = "",
  initialHealth = "on_track",
  submitLabel = "Post update",
  users,
  onSubmit,
  onCancel,
}: TProps) {
  const [body, setBody] = useState(initialBody);
  const [health, setHealth] = useState<THealth>(initialHealth);
  const [editorKey, setEditorKey] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const canSubmit = body.trim().length > 0 && !isSubmitting;

  const submit = async () => {
    if (!canSubmit) return;
    setIsSubmitting(true);
    try {
      await onSubmit({ body: body.trim(), health });
      setBody("");
      setEditorKey((key) => key + 1);
    } catch {
      return;
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="flex flex-col gap-3 p-4">
      <MarkdownEditor
        key={editorKey}
        value={body}
        onChange={setBody}
        users={users}
        placeholder="Write an update…"
        label="Update"
        onSubmit={() => void submit()}
      />
      <div className="flex items-center gap-2">
        <Select
          items={healthItems}
          value={health}
          onValueChange={(value) => value && setHealth(value as THealth)}
        >
          <SelectTrigger size="sm" className="w-36" aria-label="Health">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {HEALTHS.map((option) => (
              <SelectItem key={option} value={option}>
                <HealthIcon health={option} />
                {healthLabels[option]}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <div className="ml-auto flex items-center gap-2">
          {onCancel && (
            <Button variant="outline" size="sm" onClick={onCancel}>
              Cancel
            </Button>
          )}
          <Button size="sm" disabled={!canSubmit} onClick={() => void submit()}>
            {submitLabel}
          </Button>
        </div>
      </div>
    </Card>
  );
}
