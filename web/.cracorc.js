const isCoverage = process.env.COVERAGE === "true";
const babelPlugins = isCoverage ? ["babel-plugin-istanbul"] : [];
const cracoPlugins = isCoverage
    ? [
          {
              plugin: require("craco-alias"),
              options: {
                  source: "tsconfig",
                  baseUrl: "./src",
                  tsConfigPath: "./tsconfig.aliases.json",
              },
          },
      ]
    : [
          {
              plugin: require("craco-alias"),
              options: {
                  source: "tsconfig",
                  baseUrl: "./src",
                  tsConfigPath: "./tsconfig.aliases.json",
              },
          },
          {
              plugin: require("craco-esbuild"),
              options: {
                  enableSvgr: true,
                  esbuildMinimizerOptions: {
                      target: "es2015",
                      css: true,
                      minify: true,
                  },
                  skipEsbuildJest: true,
              },
          },
      ];

module.exports = {
    babel: {
        plugins: babelPlugins,
    },
    plugins: cracoPlugins,
};
