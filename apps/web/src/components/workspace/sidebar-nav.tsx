import { type LinkProps } from "@tanstack/react-router";
import { ChevronRightIcon, type LucideIcon } from "lucide-react";
import { ReactNode, useState } from "react";

import { Button, LinkButton } from "@/components/ui/button";

type TNavItemProps = {
  label: string;
  icon: LucideIcon;
  link: LinkProps;
  badge?: ReactNode;
  exact?: boolean;
  onNavigate?: () => void;
};

export function SidebarNavItem({
  label,
  icon: Icon,
  link,
  badge,
  exact,
  onNavigate,
}: TNavItemProps) {
  return (
    <LinkButton
      {...link}
      activeOptions={{ exact: exact ?? false, includeSearch: false }}
      variant="ghost"
      size="sm"
      forceMinSize={false}
      onClick={onNavigate}
      className="text-muted-foreground data-[status=active]:bg-accent data-[status=active]:text-foreground w-full justify-start font-normal"
    >
      <Icon className="size-4" />
      <span className="min-w-0 flex-1 truncate text-left">{label}</span>
      {badge}
    </LinkButton>
  );
}

export function SidebarSection({
  title,
  action,
  children,
  defaultOpen = true,
}: {
  title: string;
  action?: ReactNode;
  children: ReactNode;
  defaultOpen?: boolean;
}) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  return (
    <div className="flex flex-col gap-0.5">
      <div className="flex items-center gap-0.5">
        <Button
          variant="ghost"
          size="sm"
          forceMinSize={false}
          aria-expanded={isOpen}
          onClick={() => setIsOpen((open) => !open)}
          className="text-muted-foreground min-w-0 flex-1 justify-start text-xs"
        >
          <span className="truncate">{title}</span>
          <ChevronRightIcon
            data-open={isOpen || undefined}
            className="size-3.5 transition-transform data-open:rotate-90"
          />
        </Button>
        {action}
      </div>
      {isOpen && <div className="flex flex-col gap-0.5">{children}</div>}
    </div>
  );
}

export function SidebarCollapsible({
  label,
  icon,
  children,
  defaultOpen = true,
}: {
  label: string;
  icon: ReactNode;
  children: ReactNode;
  defaultOpen?: boolean;
}) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  return (
    <div className="flex flex-col gap-0.5">
      <Button
        variant="ghost"
        size="sm"
        forceMinSize={false}
        aria-expanded={isOpen}
        onClick={() => setIsOpen((open) => !open)}
        className="w-full justify-start font-normal"
      >
        {icon}
        <span className="min-w-0 flex-1 truncate text-left">{label}</span>
        <ChevronRightIcon
          data-open={isOpen || undefined}
          className="text-muted-foreground size-3.5 transition-transform data-open:rotate-90"
        />
      </Button>
      {isOpen && <div className="flex flex-col gap-0.5 pl-4">{children}</div>}
    </div>
  );
}
