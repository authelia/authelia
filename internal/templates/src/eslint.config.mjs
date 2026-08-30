import eslintReact from "@eslint-react/eslint-plugin";
import js from "@eslint/js";
import tsEslintPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import prettierPluginRecommended from "eslint-plugin-prettier/recommended";

export default [
    {
        ignores: ["dist/", "node_modules/"],
    },

    {
        files: ["**/*.{ts,tsx}"],
        ...js.configs.recommended,
    },

    {
        files: ["**/*.{ts,tsx}"],
        ...eslintReact.configs["recommended-typescript"],
    },

    {
        files: ["**/*.{ts,tsx}"],
        languageOptions: {
            parser: tsParser,
            parserOptions: {
                ecmaFeatures: {
                    jsx: true,
                },
                project: "./tsconfig.json",
            },
        },
        plugins: {
            "@typescript-eslint": tsEslintPlugin,
        },
        rules: {
            ...tsEslintPlugin.configs["eslint-recommended"].overrides[0].rules,
            ...tsEslintPlugin.configs.recommended.rules,
            "@eslint-react/no-class-component": "error",
            "@typescript-eslint/no-unused-vars": ["error", { args: "all", argsIgnorePattern: "^_" }],
            "no-unused-vars": "off",
        },
    },

    prettierPluginRecommended,
];
