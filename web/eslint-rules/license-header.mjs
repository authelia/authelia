// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Bump YEAR alongside the annual REUSE header refresh. The Go equivalent of this rule is the
// goheader linter, configured in .golangci.yml; keep the two templates in sync.
const YEAR = "2026";

const LINES = [`SPDX-FileCopyrightText: ${YEAR} Authelia`, "", "SPDX-License-Identifier: Apache-2.0"];

const HEADER = LINES.map((line) => (line === "" ? "//" : `// ${line}`)).join("\n");

/**
 * Requires the REUSE license header at the top of every linted file. Files whose licensing is
 * declared in REUSE.toml instead of inline (the vendored shadcn components, for example) are
 * excluded via the `ignores` of the config block that enables this rule, not by the rule itself.
 * @type {import("eslint").Rule.RuleModule}
 */
export default {
    create(context) {
        return {
            Program() {
                const source = context.sourceCode.getText();

                if (source.startsWith(HEADER)) {
                    return;
                }

                // Only fix when there is no header at all. A header that is merely wrong (a
                // stale year, say) needs a human, and prepending a second one would be worse.
                const fixable = !source.includes("SPDX-License-Identifier");

                const fix = (fixer) => fixer.insertTextBeforeRange([0, 0], `${HEADER}\n\n`);

                context.report({
                    fix: fixable ? fix : undefined,
                    loc: { column: 0, line: 1 },
                    messageId: "missingLicenseHeader",
                });
            },
        };
    },
    meta: {
        docs: {
            description: "require the REUSE license header at the top of the file",
        },
        fixable: "code",
        messages: {
            missingLicenseHeader: `File must start with the license header:\n${HEADER}`,
        },
        schema: [],
        type: "layout",
    },
};
