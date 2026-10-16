import { ReactNode } from "react";

import { Breadcrumb, TBreadcrumbItem } from "@/components/layout/breadcrumb";

type TProps = {
  title: string;
  icon?: ReactNode;
  description?: string;
  breadcrumb?: TBreadcrumbItem[];
  actions?: ReactNode;
  toolbar?: ReactNode;
};

export function PageHeader({ title, icon, description, breadcrumb, actions, toolbar }: TProps) {
  return (
    <header data-slot="page-header" className="flex w-full flex-col gap-2">
      {breadcrumb && breadcrumb.length > 0 && <Breadcrumb items={breadcrumb} />}
      <div className="flex min-h-9 w-full items-center gap-2">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          {icon && <span className="flex shrink-0 items-center [&_svg]:size-5">{icon}</span>}
          <h1 className="min-w-0 truncate text-xl leading-tight font-semibold">{title}</h1>
        </div>
        {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
      </div>
      {description && <p className="text-muted-foreground text-sm">{description}</p>}
      {toolbar}
    </header>
  );
}
