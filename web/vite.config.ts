import react from "@vitejs/plugin-react";
import type { OutputOptions, RollupOptions } from "rollup";
import { defineConfig, loadEnv } from "vite";
import checkerPlugin from "vite-plugin-checker";
import istanbul from "vite-plugin-istanbul";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd());
    const allowedHosts = env.VITE_ALLOWED_HOSTS ? env.VITE_ALLOWED_HOSTS.split(",") : [];
    const isCoverage = process.env.VITE_COVERAGE === "true";
    const isProduction = mode === "production" || process.env.NODE_ENV === "production";
    const sourcemap = isCoverage ? "inline" : undefined;

    const istanbulPlugin = isCoverage
        ? istanbul({
              checkProd: false,
              exclude: ["node_modules"],
              extension: [".js", ".jsx", ".ts", ".tsx"],
              forceBuildInstrument: true,
              include: "src/*",
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
                    assetFileNames: (assetInfo) => {
                        if (assetInfo.names.some((name) => name.endsWith(".css"))) {
                            return "static/css/[name].[hash].[ext]";
                        }

                        return "static/media/[name].[hash].[ext]";
                    },
                    chunkFileNames: (chunkInfo) => {
                        if (chunkInfo.name === "index") {
                            return `static/js/[name].[hash].js`;
                        } else {
                            if (chunkInfo.moduleIds.length === 0) {
                                return `static/js/[name].[hash].js`;
                            }

                            const last = chunkInfo.moduleIds.at(-1);

                            if (last?.includes("@mui/")) {
                                return `static/js/mui.[name].[hash].js`;
                            }

                            if (last) {
                                const regexp = /authelia\/web\/src\/([a-zA-Z]+)\/([a-zA-Z]+)/;
                                const match = regexp.exec(last);

                                if (match) {
                                    if (match[2] === "LoginPortal") {
                                        return `static/js/portal.[name].[hash].js`;
                                    } else if (match[2] === "ResetPassword") {
                                        return `static/js/reset-password.[name].[hash].js`;
                                    } else if (match[2] === "Settings") {
                                        if (chunkInfo.name === "SettingsRouter") {
                                            return `static/js/settings.router.[hash].js`;
                                        } else {
                                            return `static/js/settings.[name].[hash].js`;
                                        }
                                    } else {
                                        if (chunkInfo.name === "LoginLayout") {
                                            return `static/js/${match[1]}.Login.[hash].js`;
                                        }

                                        if (chunkInfo.name === "MinimalLayout") {
                                            return `static/js/${match[1]}.Minimal.[hash].js`;
                                        }

                                        return `static/js/${match[1]}.[name].[hash].js`;
                                    }
                                }
                            }

                            return `static/js/[name].[hash].js`;
                        }
                    },
                    entryFileNames: `static/js/[name].[hash].js`,
                } as OutputOptions,
            } as RollupOptions,
            sourcemap,
        },
        optimizeDeps: {
            include: ["@emotion/react", "@emotion/styled"],
        },
        plugins: [
            !isProduction &&
                checkerPlugin({
                    eslint: { lintCommand: "eslint . --ext .js,.jsx,.ts,.tsx", useFlatConfig: true },
                    typescript: true,
                }),
            istanbulPlugin,
            react(),
            svgr(),
            tsconfigPaths(),
        ],
        server: {
            allowedHosts: [
                "login.example.com",
                "adgpi0mox",
                "auth-dev.adgone.co.tz",
                "auth-dev-deep.adgone.co.tz",
                ...allowedHosts,
            ],
            host: "0.0.0.0",
            open: false,
            port: 3000,
            proxy: {
                "/api": {
                    changeOrigin: true,
                    target: "http://192.168.88.248:9010",
                },
                "/locales": {
                    changeOrigin: true,
                    target: "http://192.168.88.248:9010",
                },
            },
        },
        test: {
            coverage: {
                provider: "istanbul",
            },
            environment: "happy-dom",
            globals: true,
            setupFiles: ["src/setupTests.ts"],
        },
    };
});
