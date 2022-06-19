import mermaid from 'mermaid';

var config = {
  theme: 'default',
  fontFamily: '"Jost", -apple-system, blinkmacsystemfont, "Segoe UI", roboto, "Helvetica Neue", arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";',
};

document.addEventListener('DOMContentLoaded', () => {
  mermaid.initialize(config);
  mermaid.init(undefined, '.language-mermaid');
});
