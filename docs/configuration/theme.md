---
layout: default
title: Theme
parent: Configuration
nav_order: 15
---

# Theme

The theme section configures the theme and style Authelia uses.

## Configuration

```yaml
theme: light
```

## Options

### theme
<div markdown="1">
type: string 
{: .label .label-config .label-purple } 
default: light
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

There are currently 3 available themes for Authelia:
* light (default)
* dark
* grey

To enable automatic switching between themes, you can set `theme` to `auto`. The theme will be set to either `dark` or `light` depending on the user's system preference which is determined using media queries. To read more technical details about the media queries used, read the [MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme).
