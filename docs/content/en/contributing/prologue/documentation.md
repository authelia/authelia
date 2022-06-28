---
title: "Documentation"
description: "Information on contributing documentation to the Authelia project."
lead: "Authelia has great documentation however there are always things that can be added. This section describes the contribution process for the documentation even though it's incredibly easy."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  contributing:
    parent: "prologue"
weight: 130
toc: true
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
* [Node.js] and npm *(bundled with [Node.js])*

### Directions

The following steps will allow you to run the website on the localhost and view it live in your browser:

1. Run the following commands:

  ```bash
  git clone https://github.com/authelia/authelia.git
  cd authelia/docs
  npm run install
  npm run start
  ```

2. Visit [http://localhost:1313/](http://localhost:1313/) in your browser.
3. Modify pages to see the effects live in your browser.

## Front Matter

Most documents come with a front matter that looks similar to this:

```yaml
---
title: "A Page Title"
description: "This is a description of the page."
lead: "This is a page lead."
date: 2022-03-19T04:53:05+00:00
draft: false
menu:
  development:
    parent: "prologue"
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
