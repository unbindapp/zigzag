import { ApiError } from "@/lib/server/client.gen";

export function isApiError(error: unknown, status?: number): error is ApiError {
  if (!(error instanceof ApiError)) return false;
  return status === undefined || error.status === status;
}

export function errorText(error: unknown, fallback = "Something went wrong") {
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}
