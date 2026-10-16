import { toast } from "@/components/ui/toast";
import {
  type TDocumentLifecycleAction,
  useDocumentLifecycleMutation,
} from "@/lib/queries/documents";

const lifecycleToasts: Record<TDocumentLifecycleAction, string> = {
  archive: "Document archived",
  unarchive: "Document unarchived",
  trash: "Document moved to trash",
  restore: "Document restored",
  delete: "Document deleted",
};

export function useDocumentLifecycle() {
  const { mutateAsync, isPending } = useDocumentLifecycleMutation();
  const run = async (documentId: string, action: TDocumentLifecycleAction) => {
    await mutateAsync({ documentId, action });
    toast.add({ type: "success", title: lifecycleToasts[action] });
  };
  return { run, isPending };
}
