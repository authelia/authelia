import path from "path";

import reactRefresh from "@vitejs/plugin-react-refresh";
import { defineConfig, loadEnv } from "vite";
import eslintPlugin from "vite-plugin-eslint";
import istanbul from "vite-plugin-istanbul";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

const isCoverage = process.env.VITE_COVERAGE === "true";
const istanbulPlugin = isCoverage
    ? istanbul({
          include: "src/*",
          exclude: ["node_modules"],
          extension: [".js", ".jsx", ".ts", ".tsx"],
          requireEnv: true,
      })
    : undefined;
const sourcemap = isCoverage ? "inline" : undefined;

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, "env");

    function assetOutput(name: string | undefined) {
        if (name && name.endsWith(".css")) {
            return "static/css/[name].[hash].[ext]";
        }

        return "static/media/[name].[hash].[ext]";
    }

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

    return {
        base: "./",
        build: {
            sourcemap,
            outDir: "../internal/server/public_html",
            assetsDir: "static",
            rollupOptions: {
                output: {
                    entryFileNames: `static/js/[name].[hash].js`,
                    chunkFileNames: `static/js/[name].[hash].js`,
                    assetFileNames: ({ name }) => assetOutput(name),
                },
            },
        },
        server: {
            open: false,
            hmr: {
                clientPort: env.VITE_HMR_PORT || 3000,
            },
        },
        resolve: {
            alias: [
                {
                    find: "@components",
                    replacement: path.resolve(__dirname, "src/components"),
                },
            ],
        },
        plugins: [eslintPlugin(), htmlPlugin(), istanbulPlugin, reactRefresh(), svgr(), tsconfigPaths()],
    };
});
