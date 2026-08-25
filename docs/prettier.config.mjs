export default {
  bracketSameLine: true,
  embeddedLanguageFormatting: "off",
  endOfLine: "lf",
  experimentalTernaries: true,
  overrides: [
    {
      files: ["*.json"],
      options: {
        printWidth: 1,
      },
    },
  ],
  printWidth: 100000,
  quoteProps: "consistent",
  singleQuote: false,
  tabWidth: 2,
  trailingComma: "all",
};
