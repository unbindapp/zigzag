export function readCsrfToken(): string | null {
  const match = document.cookie.match(/(?:^|;\s*)(?:__Host-)?csrf_token=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}
