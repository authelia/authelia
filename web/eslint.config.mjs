import { fixupPluginRules } from "@eslint/compat";
import limegrassImportAlias from "@limegrass/eslint-plugin-import-alias";
import tsEslintPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import importPlugin from "eslint-plugin-import";
import perfectionist from "eslint-plugin-perfectionist";
import prettierPluginRecommended from "eslint-plugin-prettier/recommended";
import reactPlugin from "eslint-plugin-react";
import reactHooksPlugin from "eslint-plugin-react-hooks";

export default [
    {
        files: ["**/*.{js,mjs,cjs,jsx,ts,tsx}"],
        languageOptions: {
            ecmaVersion: "latest",
            sourceType: "module",
        },
        plugins: {
            "@limegrass/import-alias": limegrassImportAlias,
            import: importPlugin,
            perfectionist,
        },
        rules: {
            "import/no-webpack-loader-syntax": "error",
            "no-restricted-globals": ["error", "event", "fdescribe"],
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
            "perfectionist/sort-objects": "error",
            "perfectionist/sort-union-types": "error",
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
        files: ["**/*.{jsx,tsx}"],
        plugins: {
            react: fixupPluginRules(reactPlugin),
            "react-hooks": reactHooksPlugin,
        },
        rules: {
            ...reactPlugin.configs.recommended.rules,
            ...reactHooksPlugin.configs.recommended.rules,
            "react/jsx-pascal-case": ["warn", { allowAllCaps: true }],
            "react/prop-types": "off",
            "react/react-in-jsx-scope": "off",
        },
    },

    {
        files: ["**/*.{ts,tsx}"],
        languageOptions: {
            parser: tsParser,
            parserOptions: {
                project: "./tsconfig.json",
            },
        },
        plugins: {
            "@typescript-eslint": tsEslintPlugin,
        },
        rules: {
            "@typescript-eslint/no-unused-vars": ["error", { args: "all", argsIgnorePattern: "^_" }],
            "no-unused-vars": "off",
        },
    },

    prettierPluginRecommended,

    {
        ignores: [".pnpm-store", "build", "coverage", "html", "!**/.*.js"],
    },
];
