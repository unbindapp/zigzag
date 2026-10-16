import { type LinkProps } from "@tanstack/react-router";
import { LucideIcon } from "lucide-react";
import { ReactNode } from "react";

import { PageContainer } from "@/components/layout/page-container";
import { LinkButton } from "@/components/ui/button";

export type TSettingsTab = {
  label: string;
  icon: LucideIcon;
  link: LinkProps;
  exact?: boolean;
};

type TProps = {
  title: string;
  tabs: TSettingsTab[];
  children: ReactNode;
};

export function SettingsLayout({ title, tabs, children }: TProps) {
  return (
    <div className="flex min-h-0 w-full flex-1 flex-col lg:flex-row">
      <nav
        aria-label={title}
        className="scrollbar-hidden flex shrink-0 items-center gap-1 overflow-x-auto border-b px-4 py-2 lg:w-56 lg:flex-col lg:items-stretch lg:overflow-y-auto lg:border-r lg:border-b-0 lg:px-3 lg:py-5"
      >
        <p className="text-muted-foreground hidden px-2 pb-1 text-xs font-medium lg:block">
          {title}
        </p>
        {tabs.map((tab) => (
          <LinkButton
            key={tab.label}
            {...tab.link}
            activeOptions={{ exact: tab.exact ?? false, includeSearch: false }}
            variant="ghost"
            size="sm"
            forceMinSize={false}
            className="text-muted-foreground data-[status=active]:bg-accent data-[status=active]:text-foreground shrink-0 justify-start"
          >
            <tab.icon className="size-4" />
            <span className="truncate">{tab.label}</span>
          </LinkButton>
        ))}
      </nav>
      <PageContainer width="narrow">{children}</PageContainer>
    </div>
  );
}
