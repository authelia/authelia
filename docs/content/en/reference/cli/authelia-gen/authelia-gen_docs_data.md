---
title: "authelia-gen docs data"
description: "Reference for the authelia-gen docs data command."
lead: ""
date: 2022-09-16T14:21:05+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-authelia-gen"
weight: 915
toc: true
---

## authelia-gen docs data

Generate docs data files

```
authelia-gen docs data [flags]
```

### Options

```
  -h, --help   help for data
```

### Options inherited from parent commands

```
  -C, --cwd string                               Sets the CWD for git commands
      --dir.docs string                          The directory with the docs (default "docs")
      --dir.docs.cli-reference string            The directory to store the markdown in (default "en/reference/cli")
      --dir.docs.content string                  The directory with the docs content (default "content")
      --dir.docs.data string                     The directory with the docs data (default "data")
      --dir.locales string                       The locales directory in relation to the root (default "internal/server/locales")
  -d, --dir.root string                          The repository root (default "./")
  -X, --exclude strings                          Sets the names of excluded generators
      --file.bug-report string                   Sets the path of the bug report issue template file (default ".github/ISSUE_TEMPLATE/bug-report.yml")
      --file.commit-lint-config string           The commit lint javascript configuration file in relation to the root (default "web/.commitlintrc.js")
      --file.configuration-keys string           Sets the path of the keys file (default "internal/configuration/schema/keys.go")
      --file.docs-commit-msg-guidelines string   The commit message guidelines documentation file in relation to the root (default "docs/content/en/contributing/guidelines/commit-message.md")
      --file.docs.data.keys string               Sets the path of the docs keys file (default "configkeys.json")
      --file.docs.data.languages string          The languages docs data file in relation to the docs data folder (default "languages.json")
      --file.docs.data.misc string               The misc docs data file in relation to the docs data folder (default "misc.json")
      --file.feature-request string              Sets the path of the feature request issue template file (default ".github/ISSUE_TEMPLATE/feature-request.yml")
      --file.scripts.gen string                  Sets the path of the authelia-scripts gen file (default "cmd/authelia-scripts/cmd/gen.go")
      --file.server.generated string             Sets the path of the server generated file (default "internal/server/gen.go")
      --file.web-i18n string                     The i18n typescript configuration file in relation to the root (default "web/src/i18n/index.ts")
      --package.configuration.keys string        Sets the package name of the keys file (default "schema")
      --package.scripts.gen string               Sets the package name of the authelia-scripts gen file (default "cmd")
      --versions int                             the maximum number of minor versions to list in output templates (default 5)
```

### SEE ALSO

* [authelia-gen docs](authelia-gen_docs.md)	 - Generate docs
* [authelia-gen docs data keys](authelia-gen_docs_data_keys.md)	 - Generate the docs data file for configuration keys
* [authelia-gen docs data misc](authelia-gen_docs_data_misc.md)	 - Generate docs data file misc.json

