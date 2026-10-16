import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  type TWorkspaceRole,
  WORKSPACE_ROLES,
  workspaceRoleLabels,
} from "@/lib/queries/workspaces";

type TProps = {
  value: TWorkspaceRole;
  onChange: (role: TWorkspaceRole) => void;
  disabled?: boolean;
  label: string;
};

const roleItems = WORKSPACE_ROLES.map((role) => ({
  value: role,
  label: workspaceRoleLabels[role],
}));

export function RoleSelect({ value, onChange, disabled, label }: TProps) {
  return (
    <Select
      items={roleItems}
      value={value}
      disabled={disabled}
      onValueChange={(next) => next && onChange(next as TWorkspaceRole)}
    >
      <SelectTrigger size="sm" aria-label={label} className="w-28">
        <SelectValue />
      </SelectTrigger>
      <SelectContent align="end">
        {roleItems.map((item) => (
          <SelectItem key={item.value} value={item.value}>
            {item.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

export const roleSelectItems = roleItems;
