import { useNavigate } from "@tanstack/react-router";

import { useWorkspace } from "@/components/workspace/workspace-provider";
import { UNTITLED_DOCUMENT, useCreateDocumentMutation } from "@/lib/queries/documents";

export type TDocumentParent = { projectId?: string; initiativeId?: string };

export function useCreateDocument() {
  const { workspace } = useWorkspace();
  const navigate = useNavigate();
  const { mutateAsync, isPending } = useCreateDocumentMutation();
  const create = async (parent: TDocumentParent = {}) => {
    const res = await mutateAsync({
      workspace_id: workspace.id,
      title: UNTITLED_DOCUMENT,
      project_id: parent.projectId,
      initiative_id: parent.initiativeId,
    });
    await navigate({
      to: "/$workspace/document/$slug",
      params: { workspace: workspace.slug, slug: res.data.slug },
    });
  };
  return {
    create: (parent?: TDocumentParent) => void create(parent).catch(() => undefined),
    isPending,
  };
}
