// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
    extends: ["react-app", "plugin:import/errors", "plugin:import/warnings", "plugin:prettier/recommended", "prettier"],
    rules: {
        "@limegrass/import-alias/import-alias": "error",
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
