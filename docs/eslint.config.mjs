import css from "@eslint/css";
import json from "@eslint/json";
import markdown from "@eslint/markdown";
import prettier from "eslint-plugin-prettier";

// The @eslint/* packages below are parsers, not linters: ESLint has to build an AST before any
// rule can run, even one like prettier/prettier that works on the raw source text. None of their
// own rule sets are enabled, so Prettier remains the only thing with an opinion about formatting
// and prettier.config.mjs remains the only place it is configured. Do not replace a bare language
// reference with that plugin's `recommended` config: those bring rules of their own.
const rules = { "prettier/prettier": "error" };

export default [
  {
    ignores: ["public/", "resources/", "content/reference/cli/", "hugo_stats.json", "test-results-*.json"],
  },
  {
    files: ["**/*.md"],
    language: "markdown/gfm",
    plugins: { markdown, prettier },
    rules,
  },
  {
    files: ["**/*.json"],
    language: "json/json",
    plugins: { json, prettier },
    rules,
  },
  {
    files: ["**/*.scss"],
    language: "css/css",
    // The CSS parser rejects SCSS outright unless recoverable errors are tolerated.
    languageOptions: { tolerant: true },
    plugins: { css, prettier },
    rules,
  },
  {
    files: ["**/*.{js,cjs,mjs}"],
    plugins: { prettier },
    rules,
  },
];
