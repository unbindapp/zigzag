import { queryOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";

import { getGoClient } from "@/lib/server/client";
import type { TGoClient, TInput, TListItem } from "@/lib/server/types";

export type TTemplate = TListItem<TGoClient["templates"]["list"]>;
export type TTemplateData = TTemplate["data"];
export type TCreateTemplateInput = TInput<TGoClient["templates"]["create"]>;
export type TUpdateTemplateInput = TInput<TGoClient["templates"]["update"]>;

export const queryKeyTemplates = {
  all: () => ["templates"] as const,
  list: (workspaceId: string, teamId?: string) =>
    ["templates", "list", workspaceId, teamId ?? null] as const,
};

export const templatesListQuery = (workspaceId: string, teamId?: string) =>
  queryOptions({
    queryKey: queryKeyTemplates.list(workspaceId, teamId),
    queryFn: async () => {
      const res = await getGoClient().templates.list({
        workspace_id: workspaceId,
        team_id: teamId,
        kind: "issue",
      });
      return { templates: res.data ?? [] };
    },
  });

export async function createTemplate(input: TCreateTemplateInput) {
  const res = await getGoClient().templates.create(input);
  return { data: res.data };
}

export async function updateTemplate(input: TUpdateTemplateInput) {
  const res = await getGoClient().templates.update(input);
  return { data: res.data };
}

export async function deleteTemplate(input: { templateId: string }) {
  const res = await getGoClient().templates.delete({ template_id: input.templateId });
  return { data: res.data };
}

function useInvalidateTemplates() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeyTemplates.all() });
}

export function useCreateTemplateMutation() {
  const invalidate = useInvalidateTemplates();
  return useMutation({
    mutationFn: createTemplate,
    meta: { errorTitle: "Could not create template" },
    onSuccess: invalidate,
  });
}

export function useUpdateTemplateMutation() {
  const invalidate = useInvalidateTemplates();
  return useMutation({
    mutationFn: updateTemplate,
    meta: { errorTitle: "Could not update template" },
    onSuccess: invalidate,
  });
}

export function useDeleteTemplateMutation() {
  const invalidate = useInvalidateTemplates();
  return useMutation({
    mutationFn: deleteTemplate,
    meta: { errorTitle: "Could not delete template" },
    onSuccess: invalidate,
  });
}

export const templateNameMaxLength = 64;

export const TemplateFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Name is required.")
    .max(templateNameMaxLength, `Name should be at most ${templateNameMaxLength} characters.`),
  description: z.string().trim().max(500, "Keep the description under 500 characters."),
  title: z.string().max(255, "Keep the title under 255 characters."),
  body: z.string(),
  priority: z.string(),
  stateId: z.string(),
  assigneeId: z.string(),
  labelIds: z.array(z.string()),
  subIssues: z.array(z.string()),
});
