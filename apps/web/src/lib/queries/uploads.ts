import { useMutation } from "@tanstack/react-query";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TResponseData } from "@/lib/server/types";

export type TUpload = TResponseData<TGoClient["uploads"]["create"]>;

export async function createUpload(input: { workspaceId: string; file: File }) {
  const body = new FormData();
  body.append("workspace_id", input.workspaceId);
  body.append("file", input.file);
  const res = await getGoClient().uploads.create(body);
  return { data: res.data };
}

export function useCreateUploadMutation() {
  return useMutation({
    mutationFn: createUpload,
    meta: { errorTitle: "Upload failed" },
  });
}
