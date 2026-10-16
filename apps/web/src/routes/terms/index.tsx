import { createFileRoute } from "@tanstack/react-router";

import { LegalPage } from "@/components/legal/legal-page";

export const Route = createFileRoute("/terms/")({
  component: TermsPage,
});

function TermsPage() {
  const host = window.location.host;
  return (
    <LegalPage
      title="Terms of service"
      description={`Terms for using the Zigzag instance at ${host}.`}
      sections={[
        {
          heading: "The service",
          body: (
            <p>
              This instance is run by the organization operating {host} for its own members and
              invited collaborators. Access is granted by that organization and can be changed or
              withdrawn by it at any time.
            </p>
          ),
        },
        {
          heading: "Your account",
          body: (
            <p>
              Keep your credentials and API keys to yourself and tell the operator if you believe
              they were exposed. You are responsible for what is done with your account.
            </p>
          ),
        },
        {
          heading: "Content",
          body: (
            <p>
              What you put into the tracker stays yours. You give the operator and the other members
              of your workspaces the right to see and use it inside the service, as the product is
              designed to. Do not upload content you have no right to share.
            </p>
          ),
        },
        {
          heading: "Acceptable use",
          body: (
            <p>
              No attempts to access other workspaces or accounts, to disrupt the service, or to use
              it for anything unlawful.
            </p>
          ),
        },
        {
          heading: "Warranty and liability",
          body: (
            <p>
              The service is provided as is. The Zigzag software is open source under the MIT
              license and comes with no warranty; the operator of this instance decides on backups,
              availability and support.
            </p>
          ),
        },
      ]}
    />
  );
}
