import { useQuery } from "@tanstack/react-query";
import { KeySquareIcon, Trash2Icon } from "lucide-react";

import { describeAccess } from "@/components/api-key/helpers";
import { ListRow, ListRowText } from "@/components/layout/list";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { RelativeTime } from "@/components/shared/relative-time";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { type TApiKeyShallow, useDeleteApiKeyMutation } from "@/lib/queries/api-keys";
import { meQuery, meWorkspaces } from "@/lib/queries/me";
import { workspaceRoleLabels } from "@/lib/queries/workspaces";

export function ApiKeyRow({ apiKey }: { apiKey: TApiKeyShallow }) {
  const { data: me } = useQuery(meQuery);
  const { mutateAsync: deleteKey } = useDeleteApiKeyMutation();
  const names = Object.fromEntries(meWorkspaces(me).map((w) => [w.id, w.name]));
  const isExpired = apiKey.expires_at ? new Date(apiKey.expires_at).getTime() < Date.now() : false;
  const access = describeAccess(apiKey, names);

  return (
    <ListRow
      actions={
        <>
          <span className="text-muted-foreground hidden text-xs md:inline">
            {apiKey.last_used_at ? (
              <>
                Used <RelativeTime value={apiKey.last_used_at} />
              </>
            ) : (
              "Never used"
            )}
          </span>
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Revoke ${apiKey.name}`}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title="Revoke API key"
            description={`Anything using ${apiKey.name} stops working immediately.`}
            confirmLabel="Revoke key"
            onConfirm={() => deleteKey({ id: apiKey.id })}
          />
        </>
      }
    >
      <KeySquareIcon className="text-muted-foreground size-4" />
      <ListRowText primary={apiKey.name} secondary={`${apiKey.token_prefix}…`} />
      <SimpleTooltip content={access}>
        <Badge variant="muted">{workspaceRoleLabels[apiKey.role]}</Badge>
      </SimpleTooltip>
      {isExpired ? (
        <Badge variant="destructive">Expired</Badge>
      ) : (
        apiKey.expires_at && (
          <span className="text-muted-foreground hidden text-xs sm:inline">
            Expires <RelativeTime value={apiKey.expires_at} />
          </span>
        )
      )}
    </ListRow>
  );
}
