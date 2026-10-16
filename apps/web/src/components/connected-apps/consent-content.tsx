import { useQuery } from "@tanstack/react-query";
import {
  ArrowRightIcon,
  BoxIcon,
  CircleAlertIcon,
  GlobeIcon,
  MonitorIcon,
  ShieldCheckIcon,
} from "lucide-react";
import { useState } from "react";
import { z } from "zod";

import { roleDescription, WorkspaceAccessList } from "@/components/api-key/access-fields";
import {
  accessFormShape,
  accessItems,
  hasPickedWorkspace,
  pickWorkspaceMessage,
  type TAccess,
} from "@/components/api-key/helpers";
import { AuthShell } from "@/components/layout/auth-shell";
import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { roleSelectItems } from "@/components/shared/role-select";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { submitHandler, useAppForm } from "@/lib/hooks/use-app-form";
import {
  connectedAppClientQuery,
  type TConnectedAppClient,
  useApproveConnectedAppMutation,
  useDenyConnectedAppMutation,
} from "@/lib/queries/connected-apps";
import { meQuery } from "@/lib/queries/me";
import type { TWorkspaceRole } from "@/lib/queries/workspaces";

type TProps = {
  clientId?: string;
  redirectUri?: string;
  state?: string;
  codeChallenge?: string;
  resource?: string;
  scope?: string;
};

const title = "Authorize application";
const description = "An application is asking for access to your account.";

export function ConsentContent(props: TProps) {
  const { clientId, redirectUri, codeChallenge } = props;
  if (!clientId || !redirectUri || !codeChallenge) {
    return (
      <AuthShell title={title} description={description} width="wide">
        <ErrorCard message="This authorization request is incomplete. Start again from the application." />
      </AuthShell>
    );
  }
  return (
    <AuthShell title={title} description={description} width="wide" footer={<SignedInAs />}>
      <ConsentBody
        {...props}
        clientId={clientId}
        redirectUri={redirectUri}
        codeChallenge={codeChallenge}
      />
    </AuthShell>
  );
}

function ConsentBody(
  props: TProps & { clientId: string; redirectUri: string; codeChallenge: string },
) {
  const { data, isPending, error } = useQuery(
    connectedAppClientQuery({ clientId: props.clientId, redirectUri: props.redirectUri }),
  );
  if (isPending) return <ListSkeleton rows={3} />;
  if (!data) return <ErrorCard message={error?.message} />;
  return (
    <>
      <ClientSummary client={data.client} isOnDevice={isOnDeviceUri(props.redirectUri)} />
      <ConsentForm {...props} />
    </>
  );
}

const loopbackHosts = ["localhost", "127.0.0.1", "[::1]"];

function isOnDeviceUri(uri: string) {
  try {
    const { protocol, hostname } = new URL(uri);
    if (protocol !== "http:" && protocol !== "https:") return true;
    return loopbackHosts.includes(hostname);
  } catch {
    return false;
  }
}

function ClientSummary({
  client,
  isOnDevice,
}: {
  client: TConnectedAppClient;
  isOnDevice: boolean;
}) {
  return (
    <div className="bg-card flex w-full flex-col gap-3 rounded-lg border p-4">
      <div className="flex items-center gap-2">
        <BoxIcon className="size-4" />
        <p className="min-w-0 flex-1 truncate text-base font-semibold">{client.name}</p>
        {client.verified_brand ? (
          <Badge variant="success">
            <ShieldCheckIcon />
            Verified
          </Badge>
        ) : (
          <Badge variant="warning">
            <CircleAlertIcon />
            Unverified
          </Badge>
        )}
      </div>
      <div className="text-muted-foreground flex flex-col gap-1.5 text-sm">
        <p className="flex items-center gap-1.5">
          <GlobeIcon className="size-3.5 shrink-0" />
          {client.kind === "metadata_document" ? (
            <span>
              Published by <span className="text-foreground">{client.client_host}</span>
            </span>
          ) : (
            "Self-registered, unknown publisher"
          )}
        </p>
        <p className="flex items-center gap-1.5">
          <ArrowRightIcon className="size-3.5 shrink-0" />
          <span>
            Redirects to <span className="text-foreground">{client.redirect_host}</span>
          </span>
        </p>
        {isOnDevice && (
          <p className="text-warning flex items-center gap-1.5">
            <MonitorIcon className="size-3.5 shrink-0" />
            Redirects to an application running on your device
          </p>
        )}
        <p className="text-warning flex items-center gap-1.5">
          <CircleAlertIcon className="size-3.5 shrink-0" />
          Only approve if you started this yourself
        </p>
      </div>
    </div>
  );
}

function SignedInAs() {
  const { data: me } = useQuery(meQuery);
  if (!me) return null;
  return (
    <p>
      Signed in as <span className="text-foreground font-medium">{me.user.email}</span>
    </p>
  );
}

const formSchema = z.object(accessFormShape).refine(hasPickedWorkspace, pickWorkspaceMessage);

function ConsentForm({
  clientId,
  redirectUri,
  state,
  codeChallenge,
  resource,
  scope,
}: TProps & { clientId: string; redirectUri: string; codeChallenge: string }) {
  const [isRedirecting, setIsRedirecting] = useState(false);
  const approve = useApproveConnectedAppMutation();
  const deny = useDenyConnectedAppMutation();
  const leave = (url: string) => {
    setIsRedirecting(true);
    window.location.assign(url);
  };
  const isBusy = isRedirecting || approve.isPending || deny.isPending;

  const form = useAppForm({
    defaultValues: {
      role: "member" as TWorkspaceRole,
      access: "all" as TAccess,
      workspace_ids: [] as string[],
    },
    validators: { onSubmit: formSchema },
    onSubmit: async ({ value }) => {
      const res = await approve.mutateAsync({
        client_id: clientId,
        redirect_uri: redirectUri,
        state,
        code_challenge: codeChallenge,
        resource,
        scope,
        role: value.role,
        full_access: value.access === "all",
        workspace_ids: value.access === "all" ? [] : value.workspace_ids,
      });
      leave(res.data.redirect_url);
    },
  });

  return (
    <form
      className="bg-card flex w-full flex-col gap-4 rounded-lg border p-4"
      onSubmit={submitHandler(form)}
      noValidate
    >
      <form.AppField
        name="role"
        children={(field) => (
          <field.SelectField
            label="Role"
            description={roleDescription("application")}
            items={roleSelectItems}
          />
        )}
      />
      <form.AppField
        name="access"
        children={(field) => <field.SelectField label="Access" items={accessItems} />}
      />
      <form.Subscribe selector={(s) => s.values.access}>
        {(access) => (
          <form.Field
            name="workspace_ids"
            children={(field) => <WorkspaceAccessList field={field} access={access} />}
          />
        )}
      </form.Subscribe>
      <div className="flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          disabled={isBusy}
          isPending={deny.isPending}
          onClick={() =>
            deny.mutate(
              { client_id: clientId, redirect_uri: redirectUri, state },
              { onSuccess: (res) => leave(res.data.redirect_url) },
            )
          }
        >
          Deny
        </Button>
        <form.AppForm>
          <form.SubmitButton disabled={isBusy}>Approve</form.SubmitButton>
        </form.AppForm>
      </div>
    </form>
  );
}
