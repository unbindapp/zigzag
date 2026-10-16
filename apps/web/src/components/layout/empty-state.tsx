import { LucideIcon } from "lucide-react";
import { ReactNode } from "react";

type TProps = {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: ReactNode;
};

export function EmptyState({ icon: Icon, title, description, action }: TProps) {
  return (
    <div
      data-slot="empty-state"
      className="flex w-full flex-col items-center justify-center gap-2 px-4 py-12 text-center"
    >
      <div className="bg-muted text-muted-foreground flex size-10 items-center justify-center rounded-full">
        <Icon className="size-5" />
      </div>
      <p className="text-sm font-medium">{title}</p>
      {description && <p className="text-muted-foreground max-w-sm text-sm">{description}</p>}
      {action && <div className="flex items-center gap-2 pt-2">{action}</div>}
    </div>
  );
}
