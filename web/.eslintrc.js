module.exports = {
    "parser": "@typescript-eslint/parser",
    "parserOptions": {
        "project": "tsconfig.json"
    },
    "ignorePatterns": "build/*",
    "settings": {
        "import/resolver": {
            "typescript": {}
        }
    },
    "extends": [
        "react-app",
        "plugin:import/errors",
        "plugin:import/warnings",
        "prettier/@typescript-eslint",
        "plugin:prettier/recommended"
    ],
    "rules": {
        "import/order": [
            "error",
            {
                "groups": [
                    "builtin",
                    "external",
                    "internal"
                ],
                "pathGroups": [
                    {
                        "pattern": "react",
                        "group": "external",
                        "position": "before"
                    }
                ],
                "pathGroupsExcludedImportTypes": [
                    "react"
                ],
                "newlines-between": "always",
                "alphabetize": {
                    "order": "asc",
                    "caseInsensitive": true
                }
            }
        ]
    }
};