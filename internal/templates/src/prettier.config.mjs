export default {
    bracketSameLine: false,
    bracketSpacing: true,
    overrides: [
        {
            files: ["package.json"],
            options: {
                tabWidth: 2,
            },
        },
        {
            files: ["*.yaml", "*.yml"],
            options: {
                tabWidth: 2,
            },
        },
    ],
    printWidth: 120,
    quoteProps: "consistent",
    semi: true,
    singleQuote: false,
    tabWidth: 4,
    trailingComma: "all",
};
