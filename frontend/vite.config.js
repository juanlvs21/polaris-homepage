import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";

// Tailwind v4 se integra vía su plugin oficial de Vite (sin postcss.config.js).
// La configuración de tema vive en CSS (@theme en src/style.css).
//
// El frontend se compila DENTRO del árbol del backend (internal/server/dist)
// para que `//go:embed` lo incluya en el binario. En desarrollo, /api/* se
// proxea al backend de Go en :3000.
export default defineConfig({
  plugins: [tailwindcss()],
  build: {
    outDir: "../internal/server/dist",
    emptyOutDir: true,
    assetsDir: "static",
  },
  server: {
    port: Number(process.env.PORT) || 5173,
    proxy: {
      "/api": "http://localhost:3000",
    },
  },
});
