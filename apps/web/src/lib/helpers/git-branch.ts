const TITLE_SLUG_MAX_LENGTH = 50;

export function titleSlug(title: string) {
  return title
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, TITLE_SLUG_MAX_LENGTH)
    .replace(/-+$/g, "");
}

export function gitBranchName({
  identifier,
  title,
  displayName,
}: {
  identifier: string;
  title: string;
  displayName?: string | null;
}) {
  const slug = titleSlug(title);
  const base = slug ? `${identifier.toLowerCase()}-${slug}` : identifier.toLowerCase();
  if (!displayName) return base;
  return `${displayName}/${base}`;
}
