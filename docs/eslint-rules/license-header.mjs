// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Bump YEAR alongside the annual REUSE header refresh. The equivalents of this rule are the
// goheader linter in .golangci.yml and web/eslint-rules/license-header.mjs; keep all three in sync.
const YEAR = "2026";

const LINES = [`SPDX-FileCopyrightText: ${YEAR} Authelia`, "", "SPDX-License-Identifier: Apache-2.0"];

// Hugo content carries its REUSE header as YAML comments at the top of the frontmatter, where it
// stays out of the rendered page.
const HEADER = LINES.map((line) => (line === "" ? "#" : `# ${line}`)).join("\n");

const OPEN = "---\n";

const EXPECTED = `${OPEN}${HEADER}\n`;

/**
 * Requires the REUSE license header as the first comments of the YAML frontmatter. The generated
 * CLI reference is annotated in REUSE.toml instead of inline and is already covered by the global
 * `ignores` of the docs config, so it never reaches this rule.
 * @type {import("eslint").Rule.RuleModule}
 */
export default {
  create(context) {
    return {
      root() {
        const source = context.sourceCode.getText();

        if (source.startsWith(EXPECTED)) {
          return;
        }

        // Only fix when there is no header at all and there is frontmatter to insert it into. A
        // header that is merely wrong (a stale year, say) needs a human and prepending a second
        // one would be worse, and creating frontmatter would change how Hugo renders the page.
        const fixable = !source.includes("SPDX-License-Identifier") && source.startsWith(OPEN);

        const fix = (fixer) => fixer.insertTextAfterRange([0, OPEN.length], `${HEADER}\n\n`);

        context.report({
          fix: fixable ? fix : undefined,
          // The markdown language reports 1-based columns, unlike the JavaScript one.
          loc: { column: 1, line: 1 },
          messageId: "missingLicenseHeader",
        });
      },
    };
  },
  meta: {
    docs: {
      description: "require the REUSE license header at the top of the frontmatter",
    },
    fixable: "code",
    messages: {
      missingLicenseHeader: `Frontmatter must start with the license header:\n${EXPECTED}`,
    },
    schema: [],
    type: "layout",
  },
};
