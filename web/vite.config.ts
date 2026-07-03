import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Build-Ausgabe nach web/dist (wird vom Go-Server via embed eingebettet).
// Im Dev-Modus werden API-Aufrufe an den lokalen Server weitergereicht.
export default defineConfig({
  plugins: [react()],
  // es2022: erlaubt Top-Level-await (u.a. von noVNC genutzt). Von allen aktuellen
  // Browsern unterstützt.
  build: { outDir: "dist", emptyOutDir: true, target: "es2022" },
  server: {
    proxy: {
      "/api": { target: "http://127.0.0.1:8443", changeOrigin: true, secure: false },
    },
  },
});
