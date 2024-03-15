module.exports = {
    parser: "@typescript-eslint/parser",
    parserOptions: {
        project: "tsconfig.json",
    },
    ignorePatterns: ["build/*", "coverage/*", "!.*.js"],
    settings: {
        "import/resolver": {
            typescript: {},
        },
    },
    extends: ["react-app", "plugin:@limegrass/import-alias/recommended", "plugin:import/errors", "plugin:import/warnings", "plugin:prettier/recommended", "prettier"],
    rules: {
        "import/order": [
            "error",
            {
                groups: ["builtin", "external", "internal"],
                pathGroups: [
                    {
                        pattern: "react",
                        group: "external",
                        position: "before",
                    },
                ],
                pathGroupsExcludedImportTypes: ["react"],
                "newlines-between": "always",
                alphabetize: {
                    order: "asc",
                    caseInsensitive: true,
                },
            },
        ],
        "sort-imports": [
            "error",
            {
                ignoreCase: false,
                ignoreDeclarationSort: true,
                ignoreMemberSort: false,
                allowSeparatedGroups: false,
            },
        ],
    },
};
