import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Portal SPA. The Go backend embeds dist/ via //go:embed and serves
// at app.infrays.org/. During development, `npm run dev` proxies API
// calls to the local issuer on :8080.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    port: 5180,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
