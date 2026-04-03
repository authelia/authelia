---
title: "Documentation Contributions"
description: "Information on contributing documentation to the Authelia project."
summary: "Authelia has great documentation however there are always things that can be added. This section describes the contribution process for the documentation even though it's incredibly easy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 130
toc: true
alias:
  - /contributing/prologue/documentation
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Introduction

The website is built on [Hugo] using the [Doks] theme. [Hugo] is a powerful website building tool which allows several
simple workflows for developers as well as numerous handy features like [Shortcodes] which allow building reusable
parameterized sections of content.

## Making a Change

Anyone can simply edit the [Markdown] of the relevant document which shares a path with the website URL under the
[docs folder on GitHub]. In most if not all pages there is a link included at the very bottom which links directly to
the [Markdown] file responsible for the document.

## Viewing Changes

It's relatively easy to run the __Authelia__ website locally to test out the changes you've made.

### Requirements

* [git] *(though this can be skipped if you just download the repository)*
* [Node.js]
* [pnpm]

### Directions

The following steps will allow you to run the website on the localhost and view it live in your browser:

1. Run the following commands:
    ```bash
    git clone https://github.com/authelia/authelia.git
    cd authelia/docs
    pnpm install
    pnpm dev
    ```
2. Visit [http://localhost:1313/](http://localhost:1313/) in your browser.
3. Modify pages to see the effects live in your browser.

## Generators

There are several documentation generators that exist.

Primarily they modify the files in the following locations:

  - [docs/data](https://github.com/authelia/authelia/tree/master/docs/data) which is generated based on various changes
    throughout the repository.
  - [docs/content/reference/cli](https://github.com/authelia/authelia/tree/master/docs/content/reference/cli) which is
    generated based on the changes to the cobra commands.
  - [docs/static/schemas](https://github.com/authelia/authelia/tree/master/docs/static/schemas) which is generated based
    on changes to struct tags in [internal/configuration/schema](https://github.com/authelia/authelia/tree/master/internal/configuration/schema)
    and other struct tags within the repository.

However the generators also update the [Front Matter](#front-matter) dates of when a document was first created using
git history.

We recommend running the following command sequence after modification of the source code:

```shell
source bootstrap.sh
authelia-gen --exclude docs.date,docs.cli
```

Alternatively if you've changed the CLI or created new documents running the above command without
`authelia-gen --exclude docs.date,docs.cli`.

## Front Matter

Most documents come with a front matter that looks similar to this:

```yaml
---
title: "A Page Title"
description: "This is a description of the page."
summary: "This is a page lead."
date: 2022-03-19T04:53:05+00:00
draft: false
weight: 100
toc: true
---
```

The front matter controls several aspects about how the page is displayed and varying other aspects.

### Open Graph Protocol

First of all it's important to understand the [Open Graph Protocol]. This is a protocol developed by Meta / Facebook
which is utilized by most social media platforms to display a preview of a website. This is done by customizing special
HTML `<meta />` tags.

### Fields

This section documents each of the fields that we commonly use.

#### title

String. Configures the `<title />` element, the first `<h1 />` element, and the [Open Graph Protocol] `og:title` value.

#### description

String. Configures the and the [Open Graph Protocol] `og:description` value.

#### lead

String. Configures the first paragraph of a page which occurs directly after the [title](#title).

#### date

Timestamp. Configures the [Open Graph Protocol] `og:article:published_time` value. Also used in the [Blog](../../blog).

#### draft

Boolean. Configures the visibility of a page. If it's set to `true` it is invisible.

#### menu

Dictionary. Configures the menu linkage.

#### weight

Integer. Configures the position in the menu and the order in which pagination occurs.

#### toc

Boolean. Enables or disables the Table of Contents or `On This Page` section.

#### community

Boolean. Enables or disables the Community page header. This value only has an effect in the Integration section at this
stage.

[docs folder on GitHub]: https://github.com/authelia/authelia/tree/master/docs
[Hugo]: https://gohugo.io/
[Shortcodes]: https://gohugo.io/content-management/shortcodes/
[Doks]: https://getdoks.org/
[Markdown]: https://www.markdownguide.org/
[git]: https://git-scm.com/
[Node.js]: https://nodejs.org/en/
[Open Graph Protocol]: https://ogp.me/
[pnpm]: https://pnpm.io/installation
