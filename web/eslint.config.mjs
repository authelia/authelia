import limegrassImportAlias from "@limegrass/eslint-plugin-import-alias";
import tsParser from "@typescript-eslint/parser";
import importPlugin from "eslint-plugin-import";
import prettierPluginRecommended from "eslint-plugin-prettier/recommended";
import reactPlugin from "eslint-plugin-react";
import reactHooksPlugin from "eslint-plugin-react-hooks";

export default [
    {
        languageOptions: {
            parser: tsParser,
            parserOptions: {
                project: "./tsconfig.json",
            },
        },
        settings: {
            react: {
                version: "detect",
            },
            "import/resolver": {
                typescript: {},
            },
        },
    },

    {
        plugins: {
            react: reactPlugin,
            "react-hooks": reactHooksPlugin,
        },
        rules: {
            ...reactPlugin.configs.recommended.rules,
            ...reactHooksPlugin.configs.recommended.rules,

            "react/react-in-jsx-scope": "off",
            "react/jsx-pascal-case": ["warn", { allowAllCaps: true }],
            "react/prop-types": "off",
            "no-restricted-globals": ["error", "event", "fdescribe"],
            "import/no-webpack-loader-syntax": "error",
        },
    },

    {
        plugins: {
            "@limegrass/import-alias": limegrassImportAlias,
            import: importPlugin,
        },
        rules: {
            ...importPlugin.configs.errors.rules,
            ...importPlugin.configs.warnings.rules,
        },
    },

    {
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
    },

    prettierPluginRecommended,

    {
        ignores: ["build/*", "coverage/*", "!**/.*.js"],
    },
];
