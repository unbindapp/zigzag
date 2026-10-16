import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";

import { GoogleIcon } from "@/components/icons/google";
import { Logo } from "@/components/icons/logo";
import { Button, LinkButton } from "@/components/ui/button";
import { authProvidersQuery, googleSignInHref } from "@/lib/queries/google-auth";

const features = [
  {
    heading: "Everything a team tracks",
    body: "Workspaces and teams, issues with sub-issues, relations, labels, priorities, estimates and due dates, projects with milestones and status updates, initiatives, cycles that roll over, documents, saved views and templates.",
  },
  {
    heading: "Agents work the same issues",
    body: "An MCP server ships inside the app. Every API operation is a tool, so Claude, Codex, Cursor or any MCP client can read, file and update work with the role and workspaces you grant it, through OAuth or an API key.",
  },
  {
    heading: "Connected to GitHub",
    body: "Pull requests and branches that mention an issue link to it and move it through the workflow. Chosen repositories can sync their GitHub issues and comments both ways into a project of their own.",
  },
  {
    heading: "Bring your history",
    body: "Import a Linear workspace with its issues, archived ones included, comments, labels, projects and history. Imports are idempotent and can be dry-run first.",
  },
  {
    heading: "Inbox and email",
    body: "Assignments, mentions, comments and state changes land in an inbox and, when the operator sets up SMTP, in your email.",
  },
  {
    heading: "Yours to run",
    body: "Zigzag is open source under the MIT license: one binary, one Postgres database, no third-party services required. The organization running this instance owns its data.",
  },
];

export function LandingPage() {
  const { data: providers } = useQuery(authProvidersQuery());
  const host = window.location.host;
  const googleEnabled = providers?.providers.google.enabled ?? false;

  return (
    <main className="flex min-h-svh w-full flex-col">
      <header className="flex w-full items-center justify-between px-6 py-4">
        <Link to="/" aria-label="Zigzag home">
          <Logo />
        </Link>
        <LinkButton to="/sign-in" variant="outline" size="sm">
          Sign in
        </LinkButton>
      </header>

      <section className="flex w-full flex-col items-center gap-6 px-6 pt-16 pb-12 text-center">
        <div className="flex max-w-2xl flex-col gap-4">
          <h1 className="text-xl font-semibold">
            Self-hosted issue tracking, for people and their agents
          </h1>
          <p className="text-muted-foreground text-sm">
            Zigzag is an open source issue tracker in the shape of Linear. Teams plan and track work
            in issues, projects and cycles, and AI agents join through a built-in MCP server with
            the same permissions as everyone else. The instance at {host} is run by its own
            organization for its members and invited collaborators.
          </p>
        </div>
        <div className="flex flex-wrap items-center justify-center gap-2">
          {googleEnabled ? (
            <Button variant="outline" render={<a href={googleSignInHref(undefined)} />}>
              <GoogleIcon className="size-4" />
              Continue with Google
            </Button>
          ) : null}
          <LinkButton to="/sign-in">Sign in</LinkButton>
          <Button
            variant="ghost"
            render={
              <a href="https://github.com/zigzagapp-org/zigzag" target="_blank" rel="noreferrer" />
            }
          >
            Source on GitHub
          </Button>
        </div>
      </section>

      <section className="mx-auto grid w-full max-w-5xl gap-6 px-6 pb-16 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((feature) => (
          <div key={feature.heading} className="bg-card flex flex-col gap-2 rounded-lg border p-4">
            <h2 className="text-base font-semibold">{feature.heading}</h2>
            <p className="text-muted-foreground text-sm">{feature.body}</p>
          </div>
        ))}
      </section>

      <section className="mx-auto flex w-full max-w-2xl flex-col gap-2 px-6 pb-16 text-center">
        <h2 className="text-base font-semibold">About signing in with Google</h2>
        <p className="text-muted-foreground text-sm">
          Google sign-in only verifies your email address to sign you into an account on this
          instance. Zigzag stores your Google account identifier, your verified email and, if you
          have no avatar yet, your profile photo. It never reads your mail, contacts or files, and
          nothing is shared back to Google.
        </p>
      </section>

      <footer className="text-muted-foreground mt-auto flex w-full flex-wrap items-center justify-center gap-4 px-6 py-6 text-xs">
        <Link to="/" className="underline-offset-4 hover:underline">
          Home
        </Link>
        <Link to="/privacy" className="underline-offset-4 hover:underline">
          Privacy
        </Link>
        <Link to="/terms" className="underline-offset-4 hover:underline">
          Terms
        </Link>
        <a
          href="https://github.com/zigzagapp-org/zigzag"
          target="_blank"
          rel="noreferrer"
          className="underline-offset-4 hover:underline"
        >
          GitHub
        </a>
      </footer>
    </main>
  );
}
