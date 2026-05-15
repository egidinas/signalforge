import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { resolve } from "path";

export default defineConfig({
  plugins: [react()],
  root: resolve(__dirname, "demo"),
  publicDir: false,
  build: {
    outDir: resolve(__dirname, "dist-demo"),
    emptyOutDir: true,
  },
  server: {
    host: "0.0.0.0",
  },
});
