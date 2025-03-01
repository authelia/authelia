import TSESLint from "typescript-eslint";
import importPlugin from 'eslint-plugin-import';
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";
import { createTypeScriptImportResolver } from "eslint-import-resolver-typescript";
import ESLintConfigPrettier from "eslint-config-prettier";

export default TSESLint.config(
    ESLintConfigPrettier,
    importPlugin.flatConfigs.recommended,
    {
        ignores: [
            "build/*",
            "coverage/*",
            "!.*.js",
        ],
    },
    {
        plugins: {
            '@typescript-eslint': TSESLint.plugin,
        },
        files: ["**/*.ts", "**/*.tsx"],
        settings: {
            "import-x/resolver-next": [
                createTypeScriptImportResolver({
                    alwaysTryTypes: true,
                    project: "tsconfig.json",
                }),
            ],
            "import/resolver": {
                typescript: {}
            }
        },
        languageOptions: {
            ecmaVersion: 2020,
            sourceType: "script",
            parser: TSESLint.parser,
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
