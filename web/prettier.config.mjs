// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

export default {
    bracketSameLine: false,
    bracketSpacing: true,
    overrides: [
        {
            files: ["components.json", "package.json"],
            options: {
                tabWidth: 2,
            },
        },
    ],
    printWidth: 120,
    semi: true,
    singleQuote: false,
    tabWidth: 4,
    trailingComma: "all",
};
