import { createFileRoute } from "@tanstack/react-router";

import { LegalPage } from "@/components/legal/legal-page";

export const Route = createFileRoute("/privacy/")({
  component: PrivacyPage,
});

function PrivacyPage() {
  const host = window.location.host;
  return (
    <LegalPage
      title="Privacy policy"
      description={`How this Zigzag instance at ${host} handles your data.`}
      sections={[
        {
          heading: "Who runs this instance",
          body: (
            <p>
              Zigzag is self-hosted software. This instance is operated by the organization that
              runs {host}, which is the controller of the data stored here. Questions about your
              data go to that organization, not to the Zigzag project.
            </p>
          ),
        },
        {
          heading: "What is stored",
          body: (
            <>
              <p>
                Your account (email address, name, display name, avatar), the content you create or
                that concerns you (issues, comments, projects, documents, uploads, activity history,
                notifications), and session records needed to keep you signed in.
              </p>
              <p>
                If you sign in with Google, the instance stores your Google account identifier and
                verified email address to match you to your account. It does not receive your Google
                password and does not read anything else from your Google account.
              </p>
            </>
          ),
        },
        {
          heading: "How it is used",
          body: (
            <p>
              Only to run the tracker for the people on this instance: showing you your work,
              notifying you about it (in the app and by email when the operator has set up email),
              and keeping the service secure. Data is not sold and is not shared with third parties
              other than the services the operator has connected, such as an email relay or a GitHub
              integration.
            </p>
          ),
        },
        {
          heading: "Retention and deletion",
          body: (
            <p>
              Content stays as long as the workspace keeps it; trashed items are deleted after 30
              days. Removing your account deletes your personal data and keeps the content you
              authored attributed to a removed user. Ask the operator of this instance to remove
              your account.
            </p>
          ),
        },
        {
          heading: "Cookies",
          body: (
            <p>
              The app sets only the cookies it needs to keep you signed in and to protect against
              request forgery. There is no advertising or cross-site tracking.
            </p>
          ),
        },
      ]}
    />
  );
}
