const isCoverage = process.env.COVERAGE === 'true'
const babelPlugins = isCoverage ? [ "babel-plugin-istanbul" ] : []

module.exports = {
    babel: {
        plugins: babelPlugins,
    },
    plugins: [
        {
            plugin: require("craco-alias"),
            options: {
                source: "tsconfig",
                baseUrl: "./src",
                tsConfigPath: "./tsconfig.aliases.json",
            }
        }
    ]
};
