import tseslint from "typescript-eslint";
import typescriptEslintParser from "@typescript-eslint/parser";
import importPlugin from 'eslint-plugin-import';
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";

export default tseslint.config(
    importPlugin.flatConfigs.recommended,
    {
        ignores: [
            "build/*",
            "coverage/*",
            "!.*.js",
        ],
    },
    {
        files: ["**/*.ts", "**/*.tsx"],
        settings: {
            "import/resolver": {
                typescript: {}
            }
        },
        languageOptions: {
            ecmaVersion: 2020,
            sourceType: "module",
            parser: typescriptEslintParser,
            parserOptions: {
                project: "tsconfig.json",
                createDefaultProgram: true,
            },
        },
        extends: [
            eslintPluginPrettierRecommended
        ],
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
    }
);
