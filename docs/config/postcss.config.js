const autoprefixer = require('autoprefixer');
const purgecss = require('@fullhuman/postcss-purgecss');
const whitelister = require('purgecss-whitelister');

module.exports = {
  plugins: [
    autoprefixer(),
    purgecss({
      content: ['./hugo_stats.json'],
      extractors: [
        {
          extractor: (content) => {
            const els = JSON.parse(content).htmlElements;
            return els.tags.concat(els.classes, els.ids);
          },
          extensions: ['json']
        }
      ],
      dynamicAttributes: [
        'aria-expanded',
        'aria-selected',
        'data-bs-popper',
        'data-bs-target',
        'data-bs-theme',
        'data-dark-mode',
        'data-global-alert',
        'data-pane', // tabs.js
        'data-popper-placement',
        'data-sizes',
        'data-toggle-tab', // tabs.js
        'id',
        'size',
        'type'
      ],
      safelist: [
        'active',
        'btn-clipboard', // clipboards.js
        'clipboard', // clipboards.js
        'disabled',
        'hidden',
        'modal-backdrop', // search-modal.js
        'selected', // search-modal.js
        'show',
        'img-fluid',
        'blur-up',
        'lazyload',
        'lazyloaded',
        'alert-link',
        'container-fw ',
        'container-lg',
        'container-fluid',
        'offcanvas-backdrop',
        'figcaption',
        'dt',
        'dd',
        'showing',
        'hiding',
        'page-item',
        'page-link',
        'not-content',
        ...whitelister(['./assets/scss/**/*.scss', './node_modules/@thulite/doks-core/assets/scss/components/_code.scss', './node_modules/@thulite/doks-core/assets/scss/components/_expressive-code.scss', './node_modules/@thulite/doks-core/assets/scss/common/_syntax.scss', './node_modules/@docsearch/css/dist/modal.css'])
      ]
    })
  ]
};
