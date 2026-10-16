import { Link, type LinkProps } from "@tanstack/react-router";
import { ChevronRightIcon } from "lucide-react";
import { Fragment, ReactNode } from "react";

export type TBreadcrumbItem = {
  label: string;
  icon?: ReactNode;
  link?: LinkProps;
};

export function Breadcrumb({ items }: { items: TBreadcrumbItem[] }) {
  if (items.length === 0) return null;
  return (
    <nav aria-label="Breadcrumb" className="min-w-0">
      <ol className="text-muted-foreground flex min-w-0 items-center gap-1.5 text-sm">
        {items.map((item, index) => (
          <Fragment key={`${item.label}-${index}`}>
            {index > 0 && (
              <li aria-hidden className="flex shrink-0 items-center">
                <ChevronRightIcon className="size-3.5" />
              </li>
            )}
            <li className="flex min-w-0 items-center">
              <BreadcrumbEntry item={item} isLast={index === items.length - 1} />
            </li>
          </Fragment>
        ))}
      </ol>
    </nav>
  );
}

function BreadcrumbEntry({ item, isLast }: { item: TBreadcrumbItem; isLast: boolean }) {
  const content = (
    <>
      {item.icon}
      <span className="truncate">{item.label}</span>
    </>
  );
  if (!item.link || isLast) {
    return (
      <span
        aria-current={isLast ? "page" : undefined}
        className="data-current:text-foreground flex min-w-0 items-center gap-1.5"
        data-current={isLast || undefined}
      >
        {content}
      </span>
    );
  }
  return (
    <Link
      {...item.link}
      className="has-hover:hover:text-foreground focus-visible:ring-primary/8-10 flex min-w-0 items-center gap-1.5 rounded-sm outline-hidden focus-visible:ring-1"
    >
      {content}
    </Link>
  );
}
