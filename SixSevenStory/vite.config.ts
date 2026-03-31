import react from "@vitejs/plugin-react-swc";
import path from "path";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        proxy: {
            "/api": {
                target: "http://localhost:3000",
                changeOrigin: true,
            },
        },
        fs: {
            allow: [".."],
        },
    },
    publicDir: "public",
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
});
