import { Link } from "@tanstack/react-router";
import { ReactNode } from "react";

import { AuthShell } from "@/components/layout/auth-shell";

type TSection = { heading: string; body: ReactNode };

type TProps = {
  title: string;
  description: string;
  sections: TSection[];
};

export function LegalPage({ title, description, sections }: TProps) {
  return (
    <AuthShell title={title} description={description} width="wide">
      <div className="flex flex-col gap-6">
        {sections.map((section) => (
          <section key={section.heading} className="flex flex-col gap-2">
            <h2 className="text-base font-semibold">{section.heading}</h2>
            <div className="text-muted-foreground flex flex-col gap-2 text-sm">{section.body}</div>
          </section>
        ))}
        <p className="text-muted-foreground text-xs">
          <Link to="/sign-in" className="underline underline-offset-4">
            Back to sign in
          </Link>
        </p>
      </div>
    </AuthShell>
  );
}
