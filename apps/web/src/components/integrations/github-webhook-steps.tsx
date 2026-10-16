export function GithubWebhookSteps() {
  return (
    <ol className="text-muted-foreground flex list-decimal flex-col gap-1 pl-4 text-sm">
      <li>
        On GitHub, open a repository or organization, then Settings, Webhooks, Add webhook. Paste
        the webhook URL, pick content type application/json and paste the secret.
      </li>
      <li>
        Under events, choose Let me select individual events, then Pull requests and Pushes. For
        synced repositories add Issues and Issue comments too.
      </li>
    </ol>
  );
}
