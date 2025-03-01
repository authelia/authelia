import ESImport from 'eslint-plugin-import';
import TSESLint from "typescript-eslint";
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";
import Prettier from "eslint-config-prettier";
import react from "eslint-plugin-react";
import reactHooks from 'eslint-plugin-react-hooks';

export default TSESLint.config(
    {
        ignores: [
            "build/*",
            "coverage/*",
            "!.*.js",
        ],
    },
    Prettier,
    ESImport.flatConfigs.recommended,
    {
        plugins: {
            '@typescript-eslint': TSESLint.plugin,
            react,
            "react-hooks": reactHooks,
        },
        files: ["**/*.{ts,tsx}"],
        settings: {
            "import/resolver": {
                typescript: {}
            }
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
                    jsx: true
                }
            },
        },
        extends: [
            eslintPluginPrettierRecommended
        ],
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
    }
);
