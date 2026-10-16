import type { DisplaySettings } from "@/components/issues/display";
import { compactFilters, type ListIssuesFilters } from "@/components/issues/filters";
import type { TView, TViewDisplay } from "@/lib/queries/views";

export const defaultViewDisplay: TViewDisplay = {
  layout: "list",
  group_by: "state",
  order_by: "manual",
  show_sub_issues: true,
  show_completed: "all",
};

export function viewDisplayFrom(display: Partial<DisplaySettings> | undefined): TViewDisplay {
  if (!display) return defaultViewDisplay;
  return {
    layout: display.layout ?? defaultViewDisplay.layout,
    group_by: display.group_by ?? defaultViewDisplay.group_by,
    order_by: display.order_by ?? defaultViewDisplay.order_by,
    show_sub_issues: display.nest_sub_issues ?? defaultViewDisplay.show_sub_issues,
    show_completed: display.show_completed ?? defaultViewDisplay.show_completed,
  };
}

export function displaySettingsOf(view: TView): Partial<DisplaySettings> {
  const { display } = view;
  return {
    layout: display.layout,
    nest_sub_issues: display.show_sub_issues,
    show_completed: display.show_completed,
    ...(display.group_by ? { group_by: display.group_by as DisplaySettings["group_by"] } : {}),
    ...(display.order_by ? { order_by: display.order_by as DisplaySettings["order_by"] } : {}),
  };
}

export function viewFiltersFrom(filters: ListIssuesFilters): Record<string, unknown> {
  return compactFilters(filters);
}

export function filtersOf(view: TView): ListIssuesFilters {
  return compactFilters(view.filters as ListIssuesFilters);
}
