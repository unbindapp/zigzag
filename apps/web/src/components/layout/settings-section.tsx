import { FormEventHandler, ReactNode } from "react";

import { cn } from "@/components/ui/utils";

type TProps = {
  title: string;
  description?: string;
  actions?: ReactNode;
  footer?: ReactNode;
  variant?: "default" | "destructive";
  onSubmit?: FormEventHandler<HTMLFormElement>;
  children?: ReactNode;
};

export function SettingsSection({
  title,
  description,
  actions,
  footer,
  variant = "default",
  onSubmit,
  children,
}: TProps) {
  const className = cn(
    "bg-card flex w-full flex-col rounded-lg border",
    variant === "destructive" && "border-destructive/6-10",
  );
  const body = (
    <>
      <div className="flex w-full items-start gap-2 p-4">
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <h2
            data-variant={variant}
            className="data-[variant=destructive]:text-destructive text-base leading-tight font-semibold"
          >
            {title}
          </h2>
          {description && <p className="text-muted-foreground text-sm">{description}</p>}
        </div>
        {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
      </div>
      {children && <div className="flex w-full flex-col gap-4 px-4 pb-4">{children}</div>}
      {footer && (
        <div className="flex w-full items-center justify-end gap-2 border-t px-4 py-3">
          {footer}
        </div>
      )}
    </>
  );

  if (onSubmit) {
    return (
      <form
        data-slot="settings-section"
        className={className}
        onSubmit={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onSubmit(event);
        }}
      >
        {body}
      </form>
    );
  }
  return (
    <section data-slot="settings-section" className={className}>
      {body}
    </section>
  );
}
