---
title: "Portal Templates"
description: "Customize the Authelia login portal at runtime using external template definitions."
summary: "Learn how to ship and select Authelia portal templates without rebuilding the application."
date: 2025-02-14T00:00:00Z
draft: false
images: []
weight: 225
toc: true
seo:
  title: ""
  description: ""
  canonical: ""
  noindex: false
---

## Overview

Portal templates let you restyle the Authelia login experience without modifying or rebuilding the project.
Templates are resolved at runtime from JSON, CSS, and optional JavaScript assets that live alongside the
standard static branding directory. This keeps the core application lightweight while making it simple to
distribute bespoke themes across environments.

By default Authelia includes a single `default` template which mirrors the historical layout. Additional
templates and behaviours are discovered at runtime from the asset path configured by
[`server.asset_path`](/configuration/server/overview/#asset_path). The browser will fall back to the
`default` template whenever definitions cannot be loaded.

The runtime loader emits warnings to the browser console for malformed manifests, missing definitions, or
effects that fail to initialise, helping operators diagnose issues quickly.

## Runtime File Layout

Place template assets beneath the branding directory inside the configured `asset_path`. For Docker users the
default location is `/config/assets/static/branding`. The loader expects the following structure:

```text
static/
└── branding/
    ├── portal-template.json
    └── templates/
        ├── manifest.json
        └── <template-name>/
            ├── definition.json
            ├── config.json          # optional overrides
            └── effect.js            # optional JavaScript module
```

Each template folder name becomes the runtime `name` that end users select from the palette. Only files that
exist in this structure are served by Authelia, allowing you to add or remove templates without changing the
binary.

## Controlling the Active Template

Create `static/branding/portal-template.json` to choose the default template and optionally expose the UI
switcher:

```json
{
  "template": "nebula-bloom",
  "enableTemplateSwitcher": true
}
```

- `template` (optional): the template name to load on first visit. Unknown values fall back to the first entry
  in the manifest or the built-in `default` template.
- `enableTemplateSwitcher` (optional, default `false`): when `true`, authenticated and unauthenticated users
  can pick templates in the portal header. Their selection persists across page reloads for the session.

This file is re-read on each page load, so deploying a new default does not require restarting the service.

## Publishing the Catalogue

`static/branding/templates/manifest.json` describes the templates that should appear in the switcher:

```json
[
  {
    "name": "nebula-bloom",
    "displayName": "Nebula Bloom",
    "description": "Soft gradients with animated particles."
  },
  {
    "name": "gateway-flux",
    "displayName": "Gateway Flux",
    "description": "Split panel layout with hero messaging.",
    "interactive": "pointer"
  }
]
```

Only entries listed here are available for selection. Each object must include:

- `name`: matches the folder name inside `templates/`.
- `displayName`: the label displayed in the UI palette.
- `description`: short helper text shown in the switcher.
- `interactive` (optional): set to `"pointer"` to indicate additional interactivity; this is purely advisory for
  the UI.

## Designing Template Definitions

`definition.json` contains the baseline styling for a template. At minimum the file must provide a `style`
object following the schema used by [`@themes/portalTemplates.ts`](https://github.com/authelia/authelia/blob/master/web/src/themes/portalTemplates.ts).
Fields map to logical areas of the UI (page, root container, card, typography, buttons, form controls, and
status colours). Only the properties you specify are applied; everything else inherits from the built-in default.

Example skeleton:

```json
{
  "name": "nebula-bloom",
  "displayName": "Nebula Bloom",
  "description": "Soft gradients with animated particles.",
  "style": {
    "page": {
      "background": "linear-gradient(135deg, #121726, #331b44)"
    },
    "card": {
      "background": "rgba(14, 18, 36, 0.85)",
      "borderRadius": "16px",
      "padding": "2.25rem 2rem",
      "color": "#f0f4ff"
    }
  },
  "effect": {
    "module": "./static/branding/templates/nebula-bloom/effect.js?v=1"
  }
}
```

- `name`, `displayName`, and `description` default to the manifest values when omitted, but keeping them in sync
  is recommended for clarity.
- `effect` (optional) references a JavaScript module that can render animations in the dedicated background host.
  Version query strings are useful for cache busting during updates.

## Optional Overrides

If you need to customise a template per deployment, add `config.json` inside the template directory. Its
contents are deep-merged into the definition `style`, allowing subtle adjustments without copying the full
definition. For example:

```json
{
  "style": {
    "card": {
      "background": "rgba(10, 12, 26, 0.75)"
    }
  }
}
```

Omitting the file is perfectly valid. When merging fails due to invalid JSON, Authelia logs an error and falls
back to the base definition.

## Authoring Effect Modules

The optional `effect.js` file is loaded dynamically with standard ES module semantics. Export a `mount`
function (or default export) that receives the effect host container and the resolved template definition:

```js
export function mount({ container, definition }) {
  const canvas = document.createElement("canvas");
  container.appendChild(canvas);

  const context = canvas.getContext("2d");
  // Initialise animation, resize observers, etc.

  return () => {
    // Cleanup listeners or animation frames on template change.
  };
}
```

Any thrown errors surface in the browser console without breaking the login form. Use this hook for advanced
animations, particle systems, or to integrate third party libraries.

## Deployment Tips

- When running in Docker, mount your template catalogue into `/config/assets/static/branding`.
- Keep `manifest.json` and the template directories in sync; missing definitions are ignored with a warning.
- Cache headers served by Authelia respect your reverse proxy. Append version parameters (for example `?v=20250214`)
  to module URLs to invalidate cached effects.
- Retain the `default` template in the manifest for operators who prefer the stock layout.

With these structures in place your team can build and iterate on rich, self-contained portal designs while
tracking only minimal changes in upstream Authelia.
