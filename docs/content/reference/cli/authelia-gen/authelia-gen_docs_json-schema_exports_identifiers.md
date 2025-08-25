---
title: "authelia-gen docs json-schema exports identifiers"
description: "Reference for the authelia-gen docs json-schema exports identifiers command."
lead: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 915
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia-gen docs json-schema exports identifiers

Generate docs JSON schema for the identifiers exports

```
authelia-gen docs json-schema exports identifiers [flags]
```

### Options

```
  -h, --help   help for identifiers
```

### Options inherited from parent commands

```
  -C, --cwd string                                                 Sets the CWD for git commands
      --dir.authentication string                                  The authentication directory in relation to the root (default "internal/authentication")
      --dir.docs string                                            The directory with the docs (default "docs")
      --dir.docs.adr string                                        The directory with the ADR data (default "reference/architecture-decision-log")
      --dir.docs.cli-reference string                              The directory to store the markdown in (default "reference/cli")
      --dir.docs.content string                                    The directory with the docs content (default "content")
      --dir.docs.data string                                       The directory with the docs data (default "data")
      --dir.docs.static string                                     The directory with the docs static files (default "static")
      --dir.docs.static.json-schemas string                        The directory with the docs static JSONSchema files (default "schemas")
      --dir.locales string                                         The locales directory in relation to the root (default "internal/server/locales")
  -d, --dir.root string                                            The repository root (default "./")
      --dir.schema string                                          The schema directory in relation to the root (default "internal/configuration/schema")
      --dir.web string                                             The repository web directory in relation to the root directory (default "web")
  -X, --exclude strings                                            Sets the names of excluded generators
      --file.bug-report string                                     Sets the path of the bug report issue template file (default ".github/ISSUE_TEMPLATE/bug-report.yml")
      --file.commit-lint-config string                             The commit lint javascript configuration file in relation to the root (default ".commitlintrc.cjs")
      --file.configuration-keys string                             Sets the path of the keys file (default "internal/configuration/schema/keys.go")
      --file.docs-commit-msg-guidelines string                     The commit message guidelines documentation file in relation to the root (default "docs/content/contributing/guidelines/commit-message.md")
      --file.docs.data.keys string                                 Sets the path of the docs keys file (default "configkeys.json")
      --file.docs.data.languages string                            The languages docs data file in relation to the docs data folder (default "languages.json")
      --file.docs.data.misc string                                 The misc docs data file in relation to the docs data folder (default "misc.json")
      --file.docs.static.json-schemas.configuration string         Sets the path of the configuration JSONSchema (default "configuration")
      --file.docs.static.json-schemas.exports.identifiers string   Sets the path of the identifiers export JSONSchema (default "exports.identifiers")
      --file.docs.static.json-schemas.exports.totp string          Sets the path of the TOTP export JSONSchema (default "exports.totp")
      --file.docs.static.json-schemas.exports.webauthn string      Sets the path of the WebAuthn export JSONSchema (default "exports.webauthn")
      --file.docs.static.json-schemas.user-database string         Sets the path of the user database JSONSchema (default "user-database")
      --file.feature-request string                                Sets the path of the feature request issue template file (default ".github/ISSUE_TEMPLATE/feature-request.yml")
      --file.scripts.gen string                                    Sets the path of the authelia-scripts gen file (default "cmd/authelia-scripts/cmd/gen.go")
      --file.server.generated string                               Sets the path of the server generated file (default "internal/server/gen.go")
      --file.web.i18n string                                       The i18n typescript configuration file in relation to the web directory (default "src/i18n/index.ts")
      --file.web.package string                                    The node package configuration file in relation to the web directory (default "package.json")
      --latest                                                     Enables latest functionality with several generators like the JSON Schema generator
      --next                                                       Enables next functionality with several generators like the JSON Schema generator
      --package.configuration.keys string                          Sets the package name of the keys file (default "schema")
      --package.scripts.gen string                                 Sets the package name of the authelia-scripts gen file (default "cmd")
      --version-count int                                          the maximum number of minor versions to list in output templates (default 5)
      --versions strings                                           The versions to run the generator for, the special versions current and next are mutually exclusive
```

### SEE ALSO

* [authelia-gen docs json-schema exports](authelia-gen_docs_json-schema_exports.md)	 - Generate docs JSON schema for the various exports

