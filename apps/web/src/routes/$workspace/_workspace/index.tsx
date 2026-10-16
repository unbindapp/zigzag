import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/$workspace/_workspace/")({
  beforeLoad: ({ params }) => {
    throw redirect({ to: "/$workspace/my-issues", params });
  },
});
