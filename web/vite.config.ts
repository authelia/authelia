import { loadEnv } from "vite";
import istanbul from "vite-plugin-istanbul";
import svgr from "vite-plugin-svgr";
import { defineConfig } from "vite-react";
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

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, "env");

    function assetOutput(name: string | undefined) {
        if (name && name.endsWith(".css")) {
            return "static/css/[name].[hash].[ext]";
        }

        return "static/media/[name].[hash].[ext]";
    }

    return {
        build: {
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
        envDir: "env",
        eslint: {
            enable: true,
        },
        html: {
            injectData: {
                ...env,
            },
        },
        server: {
            open: false,
            hmr: {
                clientPort: env.VITE_HMR_PORT || 3000,
            },
        },
        plugins: [istanbulPlugin, svgr(), tsconfigPaths()],
    };
});
