import { type RegisterableHotkey, useHotkeys } from "@tanstack/react-hotkeys";
import { type ReactNode, useState } from "react";

import { applyPatch, type TIssueDraft, type TIssuePatch } from "@/components/issues/issue-draft";
import {
  AssigneePicker,
  CyclePicker,
  DueDatePicker,
  EstimatePicker,
  LabelsPicker,
  MilestonePicker,
  ParentPicker,
  PriorityPicker,
  ProjectPicker,
  StatePicker,
  SubscribersPicker,
  type TPickerAppearance,
} from "@/components/issues/pickers";
import { useCycles, useTeam } from "@/components/issues/picker-sources";
import type { TUserSummary } from "@/lib/queries/issues";
import { PropertyRow } from "@/components/layout/property-row";

export type TPropertyKey =
  | "state"
  | "priority"
  | "assignee"
  | "labels"
  | "project"
  | "milestone"
  | "cycle"
  | "estimate"
  | "due_date"
  | "parent";

type TSubscribers = {
  users: TUserSummary[];
  onChange: (added: TUserSummary[], removed: TUserSummary[]) => void;
};

type TProps = {
  teamId: string;
  issueId?: string;
  value: TIssueDraft;
  onChange: (patch: TIssuePatch) => void;
  layout?: "column" | "inline";
  shortcutsEnabled?: boolean;
  subscribers?: TSubscribers;
  disabled?: boolean;
};

const SHORTCUTS: { hotkey: RegisterableHotkey; property: TPropertyKey }[] = [
  { hotkey: "S", property: "state" },
  { hotkey: "P", property: "priority" },
  { hotkey: "A", property: "assignee" },
  { hotkey: "L", property: "labels" },
  { hotkey: "Shift+P", property: "project" },
  { hotkey: "Shift+C", property: "cycle" },
  { hotkey: "E", property: "estimate" },
  { hotkey: "D", property: "due_date" },
];

export function IssueProperties({
  teamId,
  issueId,
  value,
  onChange,
  layout = "column",
  shortcutsEnabled = false,
  subscribers,
  disabled,
}: TProps) {
  const [openProperty, setOpenProperty] = useState<TPropertyKey | null>(null);
  const team = useTeam(teamId);
  const { enabled: cyclesEnabled } = useCycles(teamId);
  const estimatesEnabled = Boolean(team && team.estimate_type !== "none");
  useHotkeys(
    SHORTCUTS.map(({ hotkey, property }) => ({
      hotkey,
      callback: () => setOpenProperty(property),
    })),
    { enabled: shortcutsEnabled && !disabled && openProperty === null, conflictBehavior: "allow" },
  );

  const appearance: TPickerAppearance = layout === "column" ? "row" : "compact";
  const control = (property: TPropertyKey) => ({
    appearance,
    disabled,
    open: openProperty === property,
    onOpenChange: (open: boolean) => setOpenProperty(open ? property : null),
  });
  const change = (patch: TIssuePatch) => onChange(patch);
  const draft = value;

  const fields: { property: TPropertyKey; label: string; node: ReactNode; hidden?: boolean }[] = [
    {
      property: "state",
      label: "Status",
      node: (
        <StatePicker
          teamId={teamId}
          state={draft.state}
          onChange={(state) => change({ state })}
          {...control("state")}
        />
      ),
    },
    {
      property: "priority",
      label: "Priority",
      node: (
        <PriorityPicker
          priority={draft.priority}
          onChange={(priority) => change({ priority })}
          {...control("priority")}
        />
      ),
    },
    {
      property: "assignee",
      label: "Assignee",
      node: (
        <AssigneePicker
          assignee={draft.assignee}
          onChange={(assignee) => change({ assignee })}
          {...control("assignee")}
        />
      ),
    },
    {
      property: "labels",
      label: "Labels",
      node: (
        <LabelsPicker
          teamId={teamId}
          labels={draft.labels}
          onChange={(labels) => change({ labels })}
          {...control("labels")}
        />
      ),
    },
    {
      property: "project",
      label: "Project",
      node: (
        <ProjectPicker
          teamId={teamId}
          project={draft.project}
          onChange={(project) => change(applyPatchDiff(draft, { project }))}
          {...control("project")}
        />
      ),
    },
    {
      property: "milestone",
      label: "Milestone",
      hidden: !draft.project,
      node: (
        <MilestonePicker
          projectId={draft.project?.id}
          milestone={draft.milestone}
          onChange={(milestone) => change({ milestone })}
          {...control("milestone")}
        />
      ),
    },
    {
      property: "cycle",
      label: "Cycle",
      hidden: !cyclesEnabled,
      node: (
        <CyclePicker
          teamId={teamId}
          cycle={draft.cycle}
          onChange={(cycle) => change({ cycle })}
          {...control("cycle")}
        />
      ),
    },
    {
      property: "estimate",
      label: "Estimate",
      hidden: !estimatesEnabled,
      node: (
        <EstimatePicker
          teamId={teamId}
          estimate={draft.estimate}
          onChange={(estimate) => change({ estimate })}
          {...control("estimate")}
        />
      ),
    },
    {
      property: "due_date",
      label: "Due date",
      node: (
        <DueDatePicker
          dueDate={draft.due_date}
          onChange={(due_date) => change({ due_date })}
          {...control("due_date")}
        />
      ),
    },
    {
      property: "parent",
      label: "Parent",
      node: (
        <ParentPicker
          issueId={issueId}
          parent={draft.parent}
          onChange={(parent) => change({ parent })}
          {...control("parent")}
        />
      ),
    },
  ];
  const visible = fields.filter((field) => !field.hidden);

  if (layout === "inline") {
    return (
      <div className="flex flex-wrap items-center gap-2">
        {visible.map((field) => (
          <span key={field.property} className="contents">
            {field.node}
          </span>
        ))}
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col">
      {visible.map((field) => (
        <PropertyRow key={field.property} label={field.label}>
          {field.node}
        </PropertyRow>
      ))}
      {subscribers && (
        <PropertyRow label="Subscribers">
          <SubscribersPicker
            subscribers={subscribers.users}
            onChange={subscribers.onChange}
            appearance="row"
            disabled={disabled}
          />
        </PropertyRow>
      )}
    </div>
  );
}

function applyPatchDiff(draft: TIssueDraft, patch: TIssuePatch): TIssuePatch {
  const next = applyPatch(draft, patch);
  if (next.milestone === draft.milestone) return patch;
  return { ...patch, milestone: next.milestone };
}
