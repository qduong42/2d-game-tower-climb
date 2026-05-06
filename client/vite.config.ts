import { defineConfig } from "vite";

export default defineConfig({
  root: ".",
  build: { outDir: "dist", emptyOutDir: true },
  test: { include: ["src/tests/**/*.test.ts"] },
});
