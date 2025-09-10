import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";
import checkerPlugin from "vite-plugin-checker";
import istanbul from "vite-plugin-istanbul";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd());
    const allowedHosts = env.VITE_ALLOWED_HOSTS ? env.VITE_ALLOWED_HOSTS.split(",") : [];
    const isCoverage = process.env.VITE_COVERAGE === "true";
    const sourcemap = isCoverage ? "inline" : undefined;

    const istanbulPlugin = isCoverage
        ? istanbul({
              checkProd: false,
              exclude: ["node_modules"],
              extension: [".js", ".jsx", ".ts", ".tsx"],
              forceBuildInstrument: true,
              include: ["src"],
              requireEnv: true,
          })
        : undefined;

    return {
        base: "./",
        build: {
            assetsDir: "static",
            emptyOutDir: true,
            outDir: "../internal/server/public_html",
            rollupOptions: {
                output: {
                    assetFileNames: ({ name }) => {
                        if (name && name.endsWith(".css")) {
                            return "static/css/[name].[hash].[ext]";
                        }

                        return "static/media/[name].[hash].[ext]";
                    },
                    chunkFileNames: (chunkInfo) => {
                        switch (chunkInfo.name) {
                            case "index":
                                return `static/js/[name].[hash].js`;
                            default:
                                if (chunkInfo.moduleIds.length === 0) {
                                    return `static/js/[name].[hash].js`;
                                }

                                const last = chunkInfo.moduleIds[chunkInfo.moduleIds.length - 1];

                                if (last.includes("@mui/")) {
                                    return `static/js/mui.[name].[hash].js`;
                                }

                                const match = last.match(/authelia\/web\/src\/([a-zA-Z]+)\/([a-zA-Z]+)/);

                                if (match) {
                                    switch (match[2]) {
                                        case "LoginPortal":
                                            return `static/js/portal.[name].[hash].js`;
                                        case "ResetPassword":
                                            return `static/js/reset-password.[name].[hash].js`;
                                        case "Settings":
                                            switch (chunkInfo.name) {
                                                case "SettingsRouter":
                                                    return `static/js/settings.router.[hash].js`;
                                                default:
                                                    return `static/js/settings.[name].[hash].js`;
                                            }
                                        default:
                                            switch (chunkInfo.name) {
                                                case "LoginLayout":
                                                    return `static/js/${match[1]}.Login.[hash].js`;
                                                case "MinimalLayout":
                                                    return `static/js/${match[1]}.Minimal.[hash].js`;
                                                default:
                                                    return `static/js/${match[1]}.[name].[hash].js`;
                                            }
                                    }
                                }

                                return `static/js/[name].[hash].js`;
                        }
                    },
                    entryFileNames: `static/js/[name].[hash].js`,
                },
            },
            sourcemap,
        },
        optimizeDeps: {
            include: ["@emotion/react", "@emotion/styled"],
        },
        server: {
            open: false,
            port: 3000,
            allowedHosts: ["login.example.com", ...allowedHosts],
        },
        test: {
            coverage: {
                include: ["src"],
                provider: "istanbul",
            },
            reporters: ["default", "html"],
            environment: "happy-dom",
            globals: true,
            onConsoleLog() {
                return false;
            },
            setupFiles: ["src/setupTests.ts"],
        },
        plugins: [
            checkerPlugin({
                eslint: { lintCommand: "eslint --format visualstudio --ext .js,.jsx,.ts,.tsx .", useFlatConfig: true },
                typescript: true,
            }),
            istanbulPlugin,
            react(),
            svgr(),
            tsconfigPaths(),
        ],
    };
});
