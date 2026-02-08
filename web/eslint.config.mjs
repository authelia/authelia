import limegrassImportAlias from "@limegrass/eslint-plugin-import-alias";
import tsParser from "@typescript-eslint/parser";
import importPlugin from "eslint-plugin-import";
import perfectionist from "eslint-plugin-perfectionist";
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
            "import/resolver": {
                typescript: {},
            },
            react: {
                version: "detect",
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

            "import/no-webpack-loader-syntax": "error",
            "no-restricted-globals": ["error", "event", "fdescribe"],
            "react/jsx-pascal-case": ["warn", { allowAllCaps: true }],
            "react/prop-types": "off",
            "react/react-in-jsx-scope": "off",
        },
    },

    {
        plugins: {
            "@limegrass/import-alias": limegrassImportAlias,
            import: importPlugin,
            perfectionist,
        },
        rules: {
            "no-restricted-imports": [
                "error",
                {
                    paths: [
                        {
                            importNames: ["default"],
                            message:
                                "Default React import is no longer required in React 17+ because JSX is automatically transformed without React in scope.",
                            name: "react",
                        },
                    ],
                },
            ],
            "no-unused-vars": ["error", { args: "all", argsIgnorePattern: "^_" }],
            "perfectionist/sort-array-includes": ["error"],
            "perfectionist/sort-imports": [
                "error",
                {
                    customGroups: [
                        {
                            elementNamePattern: ["^react$"],
                            groupName: "react",
                        },
                    ],
                    groups: ["react", ["builtin", "external"], "tsconfig-path", ["parent", "sibling", "index"]],
                    tsconfig: { rootDir: "." },
                },
            ],
            "perfectionist/sort-intersection-types": "off",
            "perfectionist/sort-named-imports": [
                "error",
                {
                    alphabet: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
                    ignoreCase: false,
                    type: "custom",
                },
            ],
            "perfectionist/sort-objects": ["error"],
            "perfectionist/sort-union-types": ["error"],
        },
    },

    prettierPluginRecommended,

    {
        ignores: [".pnpm-store", "build", "coverage", "!**/.*.js"],
    },
];
