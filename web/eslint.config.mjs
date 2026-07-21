import eslintReact from "@eslint-react/eslint-plugin";
import limegrassImportAlias from "@limegrass/eslint-plugin-import-alias";
import tsEslintPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import importPlugin from "eslint-plugin-import";
import perfectionist from "eslint-plugin-perfectionist";
import prettierPluginRecommended from "eslint-plugin-prettier/recommended";

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
        },
    },

    {
        files: ["**/*.{ts,tsx}"],
        ...eslintReact.configs["recommended-typescript"],
    },

    {
        rules: {
            "import/no-webpack-loader-syntax": "error",
            "no-restricted-globals": ["error", "event", "fdescribe"],
        },
    },

    {
        plugins: {
            "@limegrass/import-alias": limegrassImportAlias,
            "@typescript-eslint": tsEslintPlugin,
            import: importPlugin,
            perfectionist,
        },
        rules: {
            "@typescript-eslint/no-unused-vars": ["error", { args: "all", argsIgnorePattern: "^_" }],
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
            "no-unused-vars": "off",
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
        ignores: [".pnpm-store", "build", "coverage", "html", "!**/.*.js"],
    },
];
