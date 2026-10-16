import { useQuery } from "@tanstack/react-query";
import { BoxIcon, ShieldCheckIcon, Trash2Icon } from "lucide-react";

import { describeAccess } from "@/components/api-key/helpers";
import { ListRow, ListRowText } from "@/components/layout/list";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { RelativeTime } from "@/components/shared/relative-time";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SimpleTooltip } from "@/components/ui/tooltip";
import { type TConnectedApp, useRevokeConnectedAppMutation } from "@/lib/queries/connected-apps";
import { meQuery, meWorkspaces } from "@/lib/queries/me";
import { workspaceRoleLabels } from "@/lib/queries/workspaces";

export function ConnectedAppRow({ app }: { app: TConnectedApp }) {
  const { data: me } = useQuery(meQuery);
  const { mutateAsync: revoke } = useRevokeConnectedAppMutation();
  const names = Object.fromEntries(meWorkspaces(me).map((w) => [w.id, w.name]));
  const publisher = app.kind === "metadata_document" ? app.client_host : "Self-registered";
  return (
    <ListRow
      actions={
        <>
          <span className="text-muted-foreground hidden text-xs md:inline">
            {app.last_used_at ? (
              <>
                Used <RelativeTime value={app.last_used_at} />
              </>
            ) : (
              "Never used"
            )}
          </span>
          <ConfirmDialog
            trigger={
              <Button variant="ghost" size="icon" aria-label={`Revoke ${app.client_name}`}>
                <Trash2Icon className="size-4" />
              </Button>
            }
            title="Revoke access"
            description={`${app.client_name} loses access immediately and has to be authorized again.`}
            confirmLabel="Revoke"
            onConfirm={() => revoke({ id: app.id })}
          />
        </>
      }
    >
      <BoxIcon className="text-muted-foreground size-4" />
      <ListRowText primary={app.client_name} secondary={publisher} />
      {app.verified_brand && (
        <SimpleTooltip content="Verified publisher">
          <ShieldCheckIcon aria-label="Verified publisher" className="text-success size-4" />
        </SimpleTooltip>
      )}
      <SimpleTooltip content={describeAccess(app, names)}>
        <Badge variant="muted">{workspaceRoleLabels[app.role]}</Badge>
      </SimpleTooltip>
    </ListRow>
  );
}
