---
layout: default
title: Theme
parent: Configuration
nav_order: 11
---

# Server

The theme section configures the theme and style Authelia uses.

## Configuration

```yaml
theme:
  # The theme/style to display: light, dark, grey, custom
  name: light
  # The primary and secondary colors are only activated when the theme name is set to "custom".
  # The colour values need to be defined as their hex codes: #000000 to #FFFFFF are valid.
  # primary_color: "#1976d2"
  # secondary_color: "#ffffff"
```

### Custom Theme

Setting the theme name to `custom` allows a user to specify hex color codes to customise the portals look.
[Hex color codes](https://www.color-hex.com/) `#000000` to `#FFFFFF` are valid, 3-digit hex color codes are not accepted.

Example:
```yaml
theme:
  name: custom
  primary_color: "#1976d2"
  secondary_color: "#ffffff"
```