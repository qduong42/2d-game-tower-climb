import { defineConfig } from "vite";

export default defineConfig({
  root: ".",
  build: { outDir: "../cmd/server/dist", emptyOutDir: true },
  test: { include: ["src/tests/**/*.test.ts"] },
});
