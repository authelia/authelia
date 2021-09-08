module.exports = {
    rules: {
        "body-min-length": [2, "always", 20],
        "header-case": [2, "always", "lower-case"],
        "header-full-stop": [2, "never", "."],
        "header-max-length": [2, "always", 72],
        "type-enum": [2, "always", ["build", "ci", "docs", "feat", "fix", "perf", "refactor", "release", "revert", "test"]],
        "scope-enum": [2, "always", ["api", "autheliabot", "authentication", "authorization", "buildkite", "bundler", "cmd", "codecov", "commands", "configuration", "deps", "docker", "duo", "go", "golangci-lint", "handlers", "logging", "middlewares", "mocks", "models", "notification", "npm", "oidc", "regulation", "renovate", "reviewdog", "server", "session", "storage", "suites", "templates", "utils", "web"]],
    },
    defaultIgnores: true,
    helpUrl:
        "https://www.authelia.com/docs/contributing/commitmsg-guidelines.html",
};