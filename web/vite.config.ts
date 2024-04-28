import react from "@vitejs/plugin-react";
import { PluginOption, UserConfig, defineConfig } from "vite";
import checkerPlugin from "vite-plugin-checker";
import istanbul from "vite-plugin-istanbul";
import svgr from "vite-plugin-svgr";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig(({ command, mode }) => {
    const isCoverage = process.env.VITE_COVERAGE === "true";
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

    const config: UserConfig = {
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
        },
        test: {
            coverage: {
                provider: "istanbul",
            },
            environment: "happy-dom",
            globals: true,
            onConsoleLog() {
                return false;
            },
            setupFiles: ["src/setupTests.ts"],
        },
        plugins: [
            checkerPlugin({ eslint: { lintCommand: "eslint . --ext .js,.jsx,.ts,.tsx" }, typescript: true }),
            istanbulPlugin,
            react(),
            svgr(),
            tsconfigPaths(),
        ],
    };

    if (command === "serve") {
        config.plugins?.push(injectDevHomeLinkToIndex());
    }

    return config;
});

const injectDevHomeLinkToIndex: () => PluginOption = () => {
    return {
        name: "html-transform",
        transformIndexHtml(html: string) {
            const injectedHomeLinkHtml: string = `
            <a style="position:absolute; height: 30px; z-index: 99; top: 0px;
                    left: 0px; background-color: white;" href="https://home.example.com:8080">
                <b> &#127968; Demo Home</b>
            </a>`;

            const htmlArray = html.split("</body>");

            return htmlArray[0] + injectedHomeLinkHtml + "</body>" + htmlArray[1];
        },
    };
};
