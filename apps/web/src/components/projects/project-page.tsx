import { useQuery } from "@tanstack/react-query";
import { BoxIcon } from "lucide-react";
import { type ReactNode } from "react";

import { ErrorCard } from "@/components/layout/error-card";
import { ListSkeleton } from "@/components/layout/list-skeleton";
import { PageContainer } from "@/components/layout/page-container";
import { useWorkspace } from "@/components/workspace/workspace-provider";
import { isApiError } from "@/lib/server/errors";
import { type TProject, projectBySlugQuery } from "@/lib/queries/projects";
import { EntityNotFound } from "@/components/navigation/not-found-template";

export function ProjectLoader({
  slug,
  children,
}: {
  slug: string;
  children: (project: TProject) => ReactNode;
}) {
  const { workspace } = useWorkspace();
  const { data, isPending, error, refetch } = useQuery(projectBySlugQuery(workspace.id, slug));
  if (isPending) {
    return (
      <PageContainer>
        <ListSkeleton />
      </PageContainer>
    );
  }
  if (error && isApiError(error, 404)) {
    return (
      <PageContainer>
        <EntityNotFound
          Icon={BoxIcon}
          description="Project not found. It was deleted, or you cannot see it."
        />
      </PageContainer>
    );
  }
  if (error || !data.project) {
    return (
      <PageContainer>
        <ErrorCard
          message={error?.message ?? "Could not load the project"}
          onRetry={() => void refetch()}
        />
      </PageContainer>
    );
  }
  return children(data.project);
}
