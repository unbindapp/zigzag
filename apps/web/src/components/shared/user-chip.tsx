import { Avatar } from "@/components/shared/avatar";

export type TUserSummary = {
  id: string;
  name: string;
  display_name?: string | null;
  avatar_url?: string | null;
};

type TProps = {
  user: TUserSummary;
  showHandle?: boolean;
};

export function UserChip({ user, showHandle }: TProps) {
  return (
    <span className="inline-flex min-w-0 items-center gap-1.5 text-sm">
      <Avatar
        id={user.id}
        name={user.name || user.display_name || "?"}
        src={user.avatar_url}
        size="sm"
      />
      <span className="truncate">{user.name || user.display_name}</span>
      {showHandle && user.display_name && (
        <span className="text-muted-foreground truncate text-xs">@{user.display_name}</span>
      )}
    </span>
  );
}
