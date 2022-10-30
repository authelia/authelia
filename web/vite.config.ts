import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";
import eslintPlugin from "vite-plugin-eslint";
import istanbul from "vite-plugin-istanbul";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

// @ts-ignore
export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, ".");
    const isCoverage = process.env.VITE_COVERAGE === "true";
    const sourcemap = isCoverage ? "inline" : undefined;

    const htmlPlugin = () => {
        return {
            name: "html-transform",
            transformIndexHtml(html: string) {
                return html.replace(/%(.*?)%/g, function (match, p1) {
                    return env[p1];
                });
            },
        };
    };

    const istanbulPlugin = isCoverage
        ? istanbul({
              include: "src/*",
              exclude: ["node_modules"],
              extension: [".js", ".jsx", ".ts", ".tsx"],
              checkProd: false,
              forceBuildInstrument: true,
              requireEnv: true,
          })
        : undefined;

    return {
        base: "./",
        build: {
            sourcemap,
            outDir: "../internal/server/public_html",
            emptyOutDir: true,
            assetsDir: "static",
            rollupOptions: {
                output: {
                    entryFileNames: `static/js/[name].[hash].js`,
                    chunkFileNames: `static/js/[name].[hash].js`,
                    assetFileNames: ({ name }) => {
                        if (name && name.endsWith(".css")) {
                            return "static/css/[name].[hash].[ext]";
                        }

                        return "static/media/[name].[hash].[ext]";
                    },
                },
            },
        },
        server: {
            port: 3000,
            open: false,
        },
        plugins: [/* eslintPlugin({ cache: false }) */, htmlPlugin(), istanbulPlugin, react(), svgr(), tsconfigPaths()],
    };
});
