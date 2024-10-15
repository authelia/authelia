const path = require('node:path');

module.exports = {
  plugins: {
    tailwindcss: { config: path.resolve(__dirname, 'tailwind.config.ts') },
    autoprefixer: {},
  },
}
