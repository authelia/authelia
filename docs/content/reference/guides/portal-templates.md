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
            ├── style.css            # required stylesheet for the template
            └── effect.js            # optional JavaScript module
```

Each template folder name becomes the runtime `name` that end users select from the palette. The loader always
expects a `style.css` file for non-default templates so that visual adjustments live in standard stylesheets.
Only files that exist in this structure are served by Authelia, allowing you to add or remove templates without
changing the binary.

## Controlling the Active Template

Create `static/branding/portal-template.json` to choose the default template and optionally expose the UI
switcher:

```json
{
  "template": "nebula",
  "enableTemplateSwitcher": true
}
```

- `template` (optional): the template name to load on first visit. Unknown values fall back to the first entry
  in the manifest or the built-in `default` template.
- `enableTemplateSwitcher` (optional, default `false`): when `true`, authenticated and unauthenticated users
  can pick templates in the portal header. Their selection persists across page reloads for the session.

### Configuration Keys

Portal behaviour can also be managed directly in `configuration.yml`:

```yaml
portal:
  portal_template: nebula          # Optional. Name from the manifest (or `none`/`default`).
  portal_template_switcher: true   # Optional. Enables the in-portal switcher.
  portal_headline: "AndrewMohawk SSO Portal"
  portal_subtitle: "Authenticate to access homelab services"
```

- `portal_template` sets the initial template. Unknown values fall back to `default`.
- `portal_template_switcher` toggles the palette UI and persists the user’s choice via `localStorage`.
- `portal_headline` / `portal_subtitle` provide optional copy above the login form; omit them to hide the headings.

Configuration changes are applied on the next page load; a service restart is not required.

This file is re-read on each page load, so deploying a new default does not require restarting the service.

### Portal Headline Text

Set the optional top-level configuration keys `portal_headline` and `portal_subtitle` to brand the login form
directly from `configuration.yml`. When these values are empty or omitted the headings are hidden entirely,
allowing templates to present a minimal chrome or rely on custom branding within the template itself.

## Publishing the Catalogue

`static/branding/templates/manifest.json` describes the templates that should appear in the switcher and where
to fetch their assets:

```json
[
  {
    "name": "nebula",
    "displayName": "Nebula Bloom",
    "description": "Soft gradients with animated particles.",
    "definitionPath": "./static/branding/templates/nebula/definition.json",
    "stylePath": "./static/branding/templates/nebula/style.css",
    "effectPath": "./static/branding/templates/nebula/effect.js"
  },
  {
    "name": "gateway",
    "displayName": "Gateway Flux",
    "description": "Split panel layout with hero messaging.",
    "stylePath": "./static/branding/templates/gateway/style.css",
    "effectPath": "./static/branding/templates/gateway/effect.js",
    "interactive": "pointer"
  }
]
```

Only entries listed here are available for selection. Each object must include:

- `name`: matches the folder name inside `templates/`.
- `displayName`: the label displayed in the UI palette.
- `description`: short helper text shown in the switcher.
- `stylePath`: relative (or absolute) path to the stylesheet. If omitted Authelia falls back to
  `./static/branding/templates/<name>/style.css`.
- `definitionPath` (optional): override the default `definition.json` path.
- `effectPath` (optional): loads the given JavaScript module and mounts it into the effect host.
- `interactive` (optional): set to `"pointer"` to indicate additional interactivity; this is purely advisory for
  the UI.

## Designing Template Definitions

`definition.json` provides lightweight metadata and optional behavioural hints for a template. The file
inherits the schema from [`@themes/portalTemplates.ts`](https://github.com/authelia/authelia/blob/master/web/src/themes/portalTemplates.ts)
but you rarely need to populate the full `style` object now that styling lives in CSS. Most templates only
specify identification fields and `layout` preferences while deferring all visual rules to `style.css`.

Example skeleton:

```json
{
  "name": "nebula",
  "displayName": "Nebula Bloom",
  "description": "Soft gradients with animated particles.",
  "effect": {
    "module": "./static/branding/templates/nebula/effect.js?v=1"
  },
  "layout": {
    "cardVariant": "minimal",
    "maxWidth": "xs"
  }
}
```

- `name`, `displayName`, and `description` default to the manifest values when omitted, but keeping them in sync
  is recommended for clarity.
- `layout` (optional) tweaks placement defaults (card variant, container width, alignment) without relying on CSS.
- `effect` (optional) references a JavaScript module that can render animations in the dedicated background host.
  Version query strings are useful for cache busting during updates.

## Template CSS

Every template directory must include `style.css`. Authelia injects the stylesheet into the document head
whenever that template is active, using `cache: "no-store"` so you can iterate rapidly without rebuilding the
bundle. The following selectors remain stable across releases and make targeting straightforward:

- `body[data-portal-template="<name>"]`: set for every template, including the built-in `default`.
- `[data-portal-role="page"]`, `[data-portal-role="root"]`, and `[data-portal-role="card"]`: wrap the viewport,
  root grid, and primary login container.
- `[data-portal-role="content"]`: contains the logo, headings, and first-factor form.
- `.portal-template-effect`: host element for any JavaScript-driven visual effects.

Use CSS for traditional layout tweaks (split panels, decorative pseudo-elements, additional gradients) while
keeping theme tokens in `definition.json`. This hybrid approach mirrors how the example catalogue included in
the repository layers dramatic visuals on top of the structured design tokens.

Example skeleton:

```css
body[data-portal-template=\"nebula\"] {
    color: #f5f7ff;
    background: radial-gradient(circle at 30% 20%, #1b1e3d, #090b19 65%);
}

body[data-portal-template=\"nebula\"] [data-portal-role=\"card\"] {
    border-radius: 20px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    backdrop-filter: blur(24px);
    box-shadow: 0 28px 64px -36px rgba(5, 8, 20, 0.85);
}
```

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
