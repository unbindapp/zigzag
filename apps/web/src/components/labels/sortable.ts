import {
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { arrayMove, sortableKeyboardCoordinates } from "@dnd-kit/sortable";

export function useSortSensors() {
  return useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );
}

export function midpoint(before: number | undefined, after: number | undefined) {
  if (before === undefined && after === undefined) return 0;
  if (before === undefined) return (after ?? 0) - 1;
  if (after === undefined) return before + 1;
  return (before + after) / 2;
}

export function reorderedPosition<T extends { id: string; sort_order: number }>(
  items: T[],
  event: DragEndEvent,
) {
  const { active, over } = event;
  if (!over || active.id === over.id) return null;
  const ids = items.map((item) => item.id);
  const order = arrayMove(ids, ids.indexOf(String(active.id)), ids.indexOf(String(over.id)));
  const index = order.indexOf(String(active.id));
  const before = items.find((item) => item.id === order[index - 1]);
  const after = items.find((item) => item.id === order[index + 1]);
  return {
    id: String(active.id),
    order,
    sortOrder: midpoint(before?.sort_order, after?.sort_order),
  };
}

export function sortableStyle(
  transform: { x: number; y: number } | null,
  transition: string | undefined,
) {
  return {
    transform: transform ? `translate3d(${transform.x}px, ${transform.y}px, 0)` : undefined,
    transition,
  };
}
