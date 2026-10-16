import { ReactNode } from "react";

type TProps = {
  title: string;
  description?: string;
  actions?: ReactNode;
};

export function SectionTitle({ title, description, actions }: TProps) {
  return (
    <div className="flex w-full items-start gap-2">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <h2 className="text-base leading-tight font-semibold">{title}</h2>
        {description && <p className="text-muted-foreground text-sm">{description}</p>}
      </div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </div>
  );
}
