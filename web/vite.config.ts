import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import dts from "vite-plugin-dts";
import { resolve } from "path";

export default defineConfig({
  plugins: [
    react(),
    dts({ include: ["src"], tsconfigPath: "./tsconfig.lib.json" }),
  ],
  build: {
    lib: {
      entry: resolve(__dirname, "src/index.ts"),
      name: "SignalforgeWeb",
      fileName: "signalforge-web.es",
      formats: ["es"],
    },
    rollupOptions: {
      external: ["react", "react-dom", "react/jsx-runtime", "uplot"],
      output: {
        globals: { react: "React", "react-dom": "ReactDOM", uplot: "uPlot" },
      },
    },
    sourcemap: true,
  },
});
