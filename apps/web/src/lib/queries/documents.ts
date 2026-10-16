import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { LIST_REFETCH_INTERVAL_MS } from "@/lib/queries/issues";
import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem, TResponseData } from "@/lib/server/types";

export type TDocument = TResponseData<TGoClient["documents"]["get"]>;
export type TDocumentListItem = TListItem<TGoClient["documents"]["list"]>;
export type TListDocumentsInput = TInput<TGoClient["documents"]["list"]>;
export type TCreateDocumentInput = TInput<TGoClient["documents"]["create"]>;
export type TUpdateDocumentInput = TInput<TGoClient["documents"]["update"]>;

export const DOCUMENT_TITLE_MAX_LENGTH = 255;
export const UNTITLED_DOCUMENT = "Untitled";

export const queryKeyDocuments = {
  all: () => ["documents"] as const,
  lists: () => ["documents", "list"] as const,
  list: (input: TListDocumentsInput) => ["documents", "list", input] as const,
  detail: (workspaceId: string, slug: string) =>
    ["documents", "detail", workspaceId, slug] as const,
};

export const documentsListQuery = (input: TListDocumentsInput) =>
  queryOptions({
    queryKey: queryKeyDocuments.list(input),
    queryFn: async () => {
      const res = await getGoClient().documents.list(input);
      return { documents: res.data ?? [] };
    },
    refetchInterval: LIST_REFETCH_INTERVAL_MS,
  });

export const documentBySlugQuery = (workspaceId: string, slug: string) =>
  queryOptions({
    queryKey: queryKeyDocuments.detail(workspaceId, slug),
    queryFn: async () => {
      const res = await getGoClient().documents.get({ workspace_id: workspaceId, slug });
      return { document: res.data };
    },
  });

export async function createDocument(input: TCreateDocumentInput) {
  const res = await getGoClient().documents.create(input);
  return { data: res.data };
}

export async function updateDocument(input: TUpdateDocumentInput) {
  const res = await getGoClient().documents.update(input);
  return { data: res.data };
}

export type TDocumentLifecycleAction = "archive" | "unarchive" | "trash" | "restore" | "delete";

export async function changeDocumentLifecycle(input: {
  documentId: string;
  action: TDocumentLifecycleAction;
}) {
  const documents = getGoClient().documents;
  const params = { document_id: input.documentId };
  switch (input.action) {
    case "archive":
      return { data: (await documents.archive(params)).data };
    case "unarchive":
      return { data: (await documents.unarchive(params)).data };
    case "trash":
      return { data: (await documents.trash(params)).data };
    case "restore":
      return { data: (await documents.restore(params)).data };
    case "delete":
      await documents.delete(params);
      return { data: null };
  }
}

function useInvalidateDocuments() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyDocuments.all() });
}

export function useCreateDocumentMutation() {
  const invalidate = useInvalidateDocuments();
  return useMutation({
    mutationFn: createDocument,
    meta: { errorTitle: "Could not create document" },
    onSuccess: invalidate,
  });
}

export function useUpdateDocumentMutation() {
  const invalidate = useInvalidateDocuments();
  return useMutation({
    mutationFn: updateDocument,
    meta: { errorTitle: "Could not save document" },
    onSuccess: invalidate,
  });
}

export function useDocumentLifecycleMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: changeDocumentLifecycle,
    meta: { errorTitle: "Could not change the document" },
    onSuccess: (_data, variables) =>
      queryClient.invalidateQueries({
        queryKey:
          variables.action === "delete" ? queryKeyDocuments.lists() : queryKeyDocuments.all(),
      }),
  });
}
