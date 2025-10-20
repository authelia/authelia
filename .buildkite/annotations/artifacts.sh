#!/usr/bin/env bash
set -euo pipefail

artifact_links() {
  EXTENSIONS=("")

  if [[ ${1} == "checksums.sha256" ]]; then
    EXTENSIONS+=(".sig")
  elif [[ ${1} =~ \.tar.gz$ ]]; then
    EXTENSIONS+=(".cdx.json" ".spdx.json")
  fi

  for EXT in "${EXTENSIONS[@]}"; do
  	echo "      <a href=\"artifact://${1}${EXT}\">${1}${EXT}</a><br>"
  done
}

ARCH=("amd64" "arm" "arm64")
TARGETS=("${ARCH[@]}" "public_html" "checksums")
declare -A BUILDS=(
  ["linux"]="amd64 arm arm64 amd64-musl arm-musl arm64-musl"
  ["freebsd"]="amd64"
)

PREFIX="authelia"
[[ -n "${BUILDKITE_TAG:-}" ]] && PREFIX+="-${BUILDKITE_TAG}"

echo "<h4>Artifacts</h4>"
echo '<dl class="flex flex-wrap mxn1">'
for T in "${TARGETS[@]}"; do
  echo '  <div class="m1">'
  echo "    <dt>${T}</dt>"
  echo '    <dd>'


  if [[ "${T}" == "public_html" ]]; then
    artifact_links "${PREFIX}-public_html.tar.gz"
  elif [[ "${T}" == "checksums" ]]; then
    artifact_links "checksums.sha256"
  else
  	for OS in "${!BUILDS[@]}"; do
  	  for B in ${BUILDS[${OS}]}; do
  	  	case "${B}" in
  	      	"${T}"|"${T}-musl")
  	      		artifact_links "${PREFIX}-${OS}-${B}.tar.gz"
  	      		;;
  	    	esac
  	  done
  	done

  DEB_PREFIX=$(echo "${PREFIX}" | sed -e 's/v//' -e 's/-/_/' -e '/[0-9]$/s/$/-1/')
  DEB_ARCH="${T}"
  [[ "${T}" == "arm" ]] && DEB_ARCH="armhf"
  artifact_links "${DEB_PREFIX}_${DEB_ARCH}.deb"
  fi

  echo '    </dd>'
  echo '  </div>'
done
echo "</dl>"
