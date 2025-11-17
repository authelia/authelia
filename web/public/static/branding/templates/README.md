# Template Overrides

Each portal theme ships with a full set of defaults baked into the bundle. The catalogue now lives entirely in this directory, so you can add or overhaul templates without rebuilding the image.

- `manifest.json` lists the templates (name, display label, description, and optional `interactive` flag).
- `*/definition.json` contains the base palette/layout for each template.
- Optional `config.json` files still let you layer runtime overrides without touching the base definition.

## Quickstart

1. Point `assets/static/branding/portal-template.json` at the template you want to preview (and optionally enable the in-portal switcher for rapid previews):

   ```json
   {
     "template": "aurora",
     "enableTemplateSwitcher": true
   }
   ```

2. Create `assets/static/branding/templates/<template>/config.json` with just the fields you want to change:

   ```json
   {
     "style": {
       "page": {
         "background": "linear-gradient(135deg, #010615 0%, #15254a 100%)"
       },
       "card": {
         "borderRadius": "20px",
         "shadow": "0 28px 80px rgba(0, 0, 0, 0.45)"
       },
       "buttons": {
         "containedGradient": "linear-gradient(135deg, #79ffe1, #00bbff)"
       }
     }
   }
   ```

3. Reload the portal. The runtime fetches the JSON on every page load with `cache: "no-store"`, so no container restart is required.

## Notes

- Undefined values automatically fall back to the baked-in defaults.
- Invalid JSON or merge errors are logged to the browser console and the theme reverts to its default palette.
- Removing `config.json` entirely restores the stock look for that template.
