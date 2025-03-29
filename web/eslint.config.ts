import Prettier from "eslint-config-prettier";
// @ts-ignore
import ESImport from "eslint-plugin-import";
import ESPrettierRecommended from "eslint-plugin-prettier/recommended";
import ESReact from "eslint-plugin-react";
import ESReactHooks from "eslint-plugin-react-hooks";
import TSESLint from "typescript-eslint";

export default TSESLint.config(
    {
        ignores: ["build/*", "coverage/*", "!.*.js"],
    },
    Prettier,
    ESImport.flatConfigs.recommended,
    {
        plugins: {
            "@typescript-eslint": TSESLint.plugin,
            react: ESReact,
            "react-hooks": ESReactHooks,
        },
        files: ["**/*.{ts,tsx}"],
        settings: {
            "import/resolver": {
                typescript: {},
            },
        },
        languageOptions: {
            ecmaVersion: "latest",
            sourceType: "module",
            parser: TSESLint.parser,
            parserOptions: {
                project: "tsconfig.json",
                createDefaultProgram: true,
                ecmaFeatures: {
                    impliedStrict: true,
                    jsx: true,
                },
            },
        },
        extends: [ESPrettierRecommended],
        rules: {
            "react/jsx-uses-react": "error",
            "react/jsx-uses-vars": "error",
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
);
