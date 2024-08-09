const { resolve } = require('node:path');

const project = resolve(process.cwd(), './tsconfig.json');

/** @type {import('eslint').ESLint.ConfigData} */
module.exports = {
  extends: [
    "@vercel/style-guide/eslint/node",
    "@vercel/style-guide/eslint/browser",
    "@vercel/style-guide/eslint/typescript",
    "@vercel/style-guide/eslint/react",
    "@vercel/style-guide/eslint/next",
    "eslint-config-turbo",
  ]
    .map(require.resolve)
    .concat(["eslint-config-prettier"]),
  parserOptions: {
    project,
  },
  globals: {
    React: true,
    JSX: true,
  },
  settings: {
    "import/resolver": {
      typescript: {
        project,
      },
    },
  },
  ignorePatterns: ['cli/', 'cli/index.mjs', "node_modules/", "dist/"],
  rules: {
    "@next/next/no-img-element": "off",
    "@typescript-eslint/explicit-function-return-type": "off",
    "import/no-default-export": "off",
    "jsx-a11y/no-autofocus": "off",
    "no-alert": "off",
    "react/no-array-index-key": "off",
    "react/function-component-definition": [
      2,
      {
        namedComponents: "arrow-function",
        unnamedComponents: "arrow-function",
      },
    ],
    'import/no-cycle': 'off',
    'import/no-extraneous-dependencies': 'off',
    'turbo/no-undeclared-env-vars': 'off',
    'eslint-comments/require-description': 'off',
    'no-console': 'off',
  },
};
