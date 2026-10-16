import { CircleUserIcon, CornerLeftUpIcon, ListIcon } from "lucide-react";

import type { TGroupMeta } from "@/components/issues/grouping";
import { UserAvatar } from "@/components/issues/pickers";
import { CycleIcon } from "@/components/shared/cycle-icon";
import { PriorityIcon } from "@/components/shared/priority-icon";
import { ProjectIcon } from "@/components/shared/project-icon";
import { StateIcon } from "@/components/shared/state-icon";

export function GroupIcon({ meta }: { meta: TGroupMeta }) {
  switch (meta.kind) {
    case "none":
      return <ListIcon className="text-muted-foreground size-4" />;
    case "state":
      return <StateIcon type={meta.state.type} color={meta.state.color} name={meta.state.name} />;
    case "priority":
      return <PriorityIcon priority={meta.priority} />;
    case "assignee":
      return meta.user ? (
        <UserAvatar user={meta.user} size="xs" />
      ) : (
        <CircleUserIcon className="text-muted-foreground size-4" />
      );
    case "project":
      return <ProjectIcon icon={meta.project?.icon} color={meta.project?.color} />;
    case "cycle":
      return <CycleIcon status={meta.cycle ? "active" : "upcoming"} />;
    case "label":
      return (
        <span
          aria-hidden
          className="bg-muted-foreground size-2 shrink-0 rounded-full"
          style={meta.label ? { backgroundColor: meta.label.color } : undefined}
        />
      );
    case "parent":
      return <CornerLeftUpIcon className="text-muted-foreground size-4" />;
  }
}
