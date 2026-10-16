import { type LucideIcon } from "lucide-react";

import { LinkButton } from "@/components/ui/button";

type TProps = {
  code: number;
  description: string;
  buttonText: string;
  buttonTo: string;
  Icon: LucideIcon;
};

export default function NotFoundTemplate({
  code,
  description,
  buttonText,
  buttonTo,
  Icon,
}: TProps) {
  return (
    <div className="flex w-full flex-1 flex-col items-center justify-center gap-4 px-4 py-12 text-center">
      <div className="bg-muted text-muted-foreground flex size-10 items-center justify-center rounded-full">
        <Icon className="size-5" />
      </div>
      <div className="flex flex-col gap-1">
        <p className="text-muted-foreground font-mono text-xs">{code}</p>
        <p className="text-sm font-medium">{description}</p>
      </div>
      <LinkButton to={buttonTo} variant="outline">
        {buttonText}
      </LinkButton>
    </div>
  );
}

export function EntityNotFound({ Icon, description }: { Icon: LucideIcon; description: string }) {
  return (
    <NotFoundTemplate
      code={404}
      description={description}
      buttonText="Go home"
      buttonTo="/"
      Icon={Icon}
    />
  );
}
