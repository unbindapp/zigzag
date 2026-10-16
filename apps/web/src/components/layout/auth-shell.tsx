import { ReactNode } from "react";

import { Logo } from "@/components/icons/logo";

type TProps = {
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  width?: "default" | "wide";
};

export function AuthShell({ title, description, children, footer, width = "default" }: TProps) {
  return (
    <main className="flex min-h-svh w-full flex-1 flex-col items-center justify-center px-4 py-12">
      <div
        data-width={width}
        className="flex w-full max-w-sm flex-col gap-6 data-[width=wide]:max-w-xl"
      >
        <div className="flex flex-col items-center gap-4 text-center">
          <Logo />
          <div className="flex flex-col gap-1">
            <h1 className="text-xl leading-tight font-semibold">{title}</h1>
            {description && <p className="text-muted-foreground text-sm">{description}</p>}
          </div>
        </div>
        {children}
        {footer && (
          <div className="text-muted-foreground flex flex-col items-center gap-2 text-center text-sm">
            {footer}
          </div>
        )}
      </div>
    </main>
  );
}
