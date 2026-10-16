import { defineConfig } from "vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const apiTarget = process.env.VITE_DEV_API_PROXY ?? "http://localhost:8091";

const rootApiPaths = [
  "/oauth/authorize",
  "/oauth/token",
  "/oauth/register",
  "/.well-known",
  "/mcp",
];
const rootApiProxy = Object.fromEntries(
  rootApiPaths.map((path) => [
    path,
    { target: apiTarget, changeOrigin: true, cookieDomainRewrite: "" },
  ]),
);

export default defineConfig({
  resolve: { tsconfigPaths: true },
  build: { reportCompressedSize: false },
  server: {
    port: 5174,
    strictPort: true,
    proxy: {
      "/api/go": {
        target: apiTarget,
        changeOrigin: true,
        cookieDomainRewrite: "",
        rewrite: (path) => path.replace(/^\/api\/go/, ""),
      },
      ...rootApiProxy,
    },
  },
  plugins: [tanstackRouter({ target: "react", autoCodeSplitting: true }), react(), tailwindcss()],
});
