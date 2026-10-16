import { ReactNode } from "react";

import { PageHeader } from "@/components/layout/page-header";
import { TBreadcrumbItem } from "@/components/layout/breadcrumb";

type TProps = {
  title: string;
  description?: string;
  breadcrumb?: TBreadcrumbItem[];
  actions?: ReactNode;
  children: ReactNode;
};

export function SettingsPage({ title, description, breadcrumb, actions, children }: TProps) {
  return (
    <div data-slot="settings-page" className="flex w-full flex-col gap-6">
      <PageHeader
        title={title}
        description={description}
        breadcrumb={breadcrumb}
        actions={actions}
      />
      {children}
    </div>
  );
}
