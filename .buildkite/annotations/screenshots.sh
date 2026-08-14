#!/usr/bin/env bash
set -euo pipefail

# Emits an annotation body embedding each failure screenshot captured by the suite. The artifact://
# references resolve once the artifact upload completes, so this may be emitted before it.

SUITE="${SUITE:-unknown}"

echo "<h4>${SUITE}</h4>"
echo '<dl class="flex flex-wrap mxn1">'

while IFS= read -r SCREENSHOT; do
  NAME="${SCREENSHOT##*/}"

  echo '  <div class="m1">'
  echo "    <dt>${NAME%.png}</dt>"
  echo '    <dd>'
  echo "      <a href=\"artifact://${SCREENSHOT}\"><img src=\"artifact://${SCREENSHOT}\" alt=\"${NAME}\" height=\"400\"></a>"
  echo '    </dd>'
  echo '  </div>'
done < <(find screenshots -type f -name '*.png' | sort)

echo '</dl>'
