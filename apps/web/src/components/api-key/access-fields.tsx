import type { AnyFieldApi } from "@tanstack/react-form";
import { useQuery } from "@tanstack/react-query";

import type { TAccess } from "@/components/api-key/helpers";
import { CheckboxList } from "@/components/shared/checkbox-list";
import { WorkspaceIcon } from "@/components/workspace/workspace-icon";
import { useFieldError } from "@/lib/hooks/use-app-form";
import { meQuery, meWorkspaces } from "@/lib/queries/me";

export function WorkspaceAccessList({ field, access }: { field: AnyFieldApi; access: TAccess }) {
  const { data: me } = useQuery(meQuery);
  const { error } = useFieldError(field);
  if (access !== "specific") return null;
  return (
    <div className="flex flex-col gap-1.5">
      <p className="text-sm font-medium">Workspaces</p>
      <CheckboxList
        label="Workspaces"
        items={meWorkspaces(me).map((workspace) => ({
          value: workspace.id,
          label: workspace.name,
          icon: (
            <WorkspaceIcon
              id={workspace.id}
              name={workspace.name}
              icon={workspace.icon}
              color={workspace.color}
            />
          ),
        }))}
        value={field.state.value as string[]}
        onChange={(value) => {
          field.handleChange(value);
          field.handleBlur();
        }}
      />
      {error && (
        <p role="alert" className="text-destructive text-xs">
          {error}
        </p>
      )}
    </div>
  );
}

export function roleDescription(subject: "key" | "application") {
  return `The ${subject} can never do more than you can in a workspace.`;
}
