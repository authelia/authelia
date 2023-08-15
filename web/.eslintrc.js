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
    plugins: ["@limegrass/import-alias"],
    extends: ["react-app", "plugin:prettier/recommended", "prettier"],
    rules: {
        "@limegrass/import-alias/import-alias": "error",
        "import/no-named-as-default": "warn",
        "import/no-named-as-default-member": "warn",
        "import/no-duplicates": "warn",
        "import/no-unresolved": "error",
        "import/named": "error",
        "import/namespace": "error",
        "import/default": "error",
        "import/export": "error",
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
