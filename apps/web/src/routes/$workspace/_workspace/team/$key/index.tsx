import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/$workspace/_workspace/team/$key/")({
  beforeLoad: ({ params }) => {
    throw redirect({ to: "/$workspace/team/$key/issues", params });
  },
});
