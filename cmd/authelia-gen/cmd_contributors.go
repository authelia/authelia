// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newContributorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseContributors,
		Short: "Generate the contributors card",
		RunE:  contributorsRunE,

		DisableAutoGenTag: true,
	}

	return cmd
}

func contributorsRunE(cmd *cobra.Command, args []string) (err error) {
	var root, pathDocsStatic string

	if root, err = getPFlagPath(cmd.Flags(), cmdFlagRoot); err != nil {
		return err
	}

	if pathDocsStatic, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsStatic); err != nil {
		return err
	}

	var contributors []Contributor

	if contributors, err = getContributors(filepath.Join(root, fileAllContributors)); err != nil {
		return err
	}

	var avatars []string

	if avatars, err = getContributorAvatars(contributors); err != nil {
		return err
	}

	dir := filepath.Join(pathDocsStatic, dirDocsStaticImages, dirDocsStaticImagesContributors)

	if err = os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	for i, contributor := range contributors {
		cell := getContributorCell(contributor, avatars[i])

		if isDocumentationOverlord(contributor) {
			cell = getContributorAvatarCell(contributor, avatars[i])
		}

		if err = os.WriteFile(filepath.Join(dir, contributor.ID+extSVG), []byte(cell), 0600); err != nil {
			return err
		}
	}

	path := filepath.Join(root, fileREADME)

	if err = writeContributorsCard(path, getContributorsCard(contributors)); err != nil {
		return err
	}

	fmt.Printf("Generated the card in %s with %d contributors and their cells in %s.\n", path, len(contributors), dir)

	return nil
}

func getContributors(path string) (contributors []Contributor, err error) {
	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return nil, err
	}

	config := &AllContributors{}

	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	for i, contributor := range config.Contributors {
		if config.Contributors[i].ID, err = getContributorID(contributor.AvatarURL); err != nil {
			return nil, fmt.Errorf("error occurred validating contributor '%s': %w", contributor.Login, err)
		}

		if err = validateProfileURL(contributor.Profile); err != nil {
			return nil, fmt.Errorf("error occurred validating contributor '%s': %w", contributor.Login, err)
		}
	}

	return config.Contributors, nil
}

// getContributorID returns the account an avatar url names, having checked the url is one the
// generator should fetch at all. A cell is named after the account rather than the login behind it
// because a login can be changed by the person holding it, which would leave the README asking for a
// file the generator no longer writes while the one it replaced lingered beside it. The id is also
// what makes the name safe: the roster is a file in the repository, so a login would be input rather
// than a trusted value on its way to becoming a path on disk and a segment in a URL.
func getContributorID(raw string) (id string, err error) {
	if err = validateAvatarURL(raw); err != nil {
		return "", err
	}

	var parsed *url.URL

	if parsed, err = url.Parse(raw); err != nil {
		return "", err
	}

	if id = strings.TrimPrefix(parsed.Path, contributorsAvatarPath); id == parsed.Path || id == "" {
		return "", fmt.Errorf("avatar url path must be '%s' followed by an account but is '%s'", contributorsAvatarPath, parsed.Path)
	}

	for _, r := range id {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("avatar url account must be numeric but is '%s'", id)
		}
	}

	return id, nil
}

// validateAvatarURL rejects an avatar the generator should not fetch. The roster is a file in the
// repository, so the URLs it holds are input rather than trusted values, and an avatar is fetched
// during generation.
func validateAvatarURL(raw string) (err error) {
	var parsed *url.URL

	if parsed, err = url.Parse(raw); err != nil {
		return fmt.Errorf("avatar url could not be parsed: %w", err)
	}

	if parsed.Scheme != schemeHTTPS || parsed.Host != contributorsAvatarHost {
		return fmt.Errorf("avatar url must be %s on %s but is '%s'", schemeHTTPS, contributorsAvatarHost, raw)
	}

	return nil
}

// validateProfileURL rejects a profile the generator should not link. A profile is a contributor
// supplied value which becomes an anchor in the README, so a scheme which executes rather than
// navigates must not reach it.
func validateProfileURL(raw string) (err error) {
	var parsed *url.URL

	if parsed, err = url.Parse(raw); err != nil {
		return fmt.Errorf("profile url could not be parsed: %w", err)
	}

	if parsed.Scheme != schemeHTTP && parsed.Scheme != schemeHTTPS {
		return fmt.Errorf("profile url must be %s or %s but is '%s'", schemeHTTP, schemeHTTPS, raw)
	}

	return nil
}

// getContributorAvatars fetches each avatar at the density a cell draws it, flattens any transparency
// onto white as JPEG has no alpha channel, and re-encodes it. Re-encoding is what keeps a cell near a
// third of the size it would be with the image as GitHub serves it.
func getContributorAvatars(contributors []Contributor) (avatars []string, err error) {
	avatars = make([]string, len(contributors))

	client := &http.Client{
		Timeout: time.Second * 30,
		CheckRedirect: func(req *http.Request, via []*http.Request) (err error) {
			if len(via) >= contributorsAvatarRedirects {
				return fmt.Errorf("stopped after %d redirects", contributorsAvatarRedirects)
			}

			return validateAvatarURL(req.URL.String())
		},
	}

	group := &errgroup.Group{}
	group.SetLimit(contributorsAvatarConcurrency)

	for i, contributor := range contributors {
		group.Go(func() (err error) {
			var avatar string

			if avatar, err = getContributorAvatar(client, contributor.AvatarURL); err != nil {
				return fmt.Errorf("error occurred fetching avatar for contributor '%s': %w", contributor.Login, err)
			}

			avatars[i] = avatar

			return nil
		})
	}

	if err = group.Wait(); err != nil {
		return nil, err
	}

	return avatars, nil
}

func getContributorAvatar(client *http.Client, url string) (avatar string, err error) {
	url = fmt.Sprintf("%s&s=%d", strings.Split(url, "&s=")[0], contributorsAvatarSize)

	var resp *http.Response

	if resp, err = client.Get(url); err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received status code %d", resp.StatusCode)
	}

	var (
		data []byte
		img  image.Image
	)

	if data, err = io.ReadAll(resp.Body); err != nil {
		return "", err
	}

	if img, _, err = image.Decode(bytes.NewReader(data)); err != nil {
		return "", err
	}

	bounds := img.Bounds()

	flat := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	draw.Draw(flat, flat.Bounds(), image.NewUniform(image.White), image.Point{}, draw.Src)
	draw.Draw(flat, flat.Bounds(), img, bounds.Min, draw.Over)

	buf := &bytes.Buffer{}

	if err = jpeg.Encode(buf, downscale(flat, contributorsAvatarSize), &jpeg.Options{Quality: contributorsAvatarQuality}); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// downscale reduces an avatar to fit within max pixels on its longest side by averaging each source
// box into a destination pixel. GitHub ignores the requested size for some avatars and answers with
// the full sized original, which would otherwise be embedded at many times the density a cell draws.
// An avatar already within the bound is returned untouched rather than enlarged.
func downscale(src *image.RGBA, max int) *image.RGBA {
	bounds := src.Bounds()

	width, height := bounds.Dx(), bounds.Dy()

	if width <= max && height <= max {
		return src
	}

	if width >= height {
		width, height = max, height*max/width
	} else {
		width, height = width*max/height, max
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		y0, y1 := bounds.Min.Y+y*bounds.Dy()/height, bounds.Min.Y+(y+1)*bounds.Dy()/height

		for x := range width {
			x0, x1 := bounds.Min.X+x*bounds.Dx()/width, bounds.Min.X+(x+1)*bounds.Dx()/width

			var r, g, b, n uint32

			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					sr, sg, sb, _ := src.At(sx, sy).RGBA()
					r, g, b, n = r+sr>>8, g+sg>>8, b+sb>>8, n+1
				}
			}

			dst.Set(x, y, color.RGBA{R: average(r, n), G: average(g, n), B: average(b, n), A: math.MaxUint8})
		}
	}

	return dst
}

// getContributorCell draws one contributor as a standalone SVG. GitHub strips the CSS a card would
// need and leaves nothing which rounds an avatar or sets a typeface, so each cell is drawn here and
// referenced from the README as an image instead. An SVG behind an <img> element is inert, which is
// what broke the card this replaces, but a cell needs no link of its own: the README wraps each one
// in the anchor, and the cell is left to do nothing but draw.
func getContributorCell(contributor Contributor, avatar string) string {
	center := float64(contributorsCellWidth) / 2

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">
  <style>
    svg { color-scheme: light dark }
    .muted { fill: #71717a } .name { fill: #3f3f46 }
    .ring { stroke: rgba(24, 24, 27, 0.12) }
    @media (prefers-color-scheme: dark) {
      .muted { fill: #a1a1aa } .name { fill: #d4d4d8 }
      .ring { stroke: rgba(250, 250, 250, 0.14) }
    }
  </style>
  <defs>
    <clipPath id="avatar" clipPathUnits="objectBoundingBox"><circle cx="0.5" cy="0.5" r="0.5" /></clipPath>
  </defs>
  <image href="data:image/jpeg;base64,%s" x="%.1f" y="0" width="%d" height="%d" preserveAspectRatio="xMidYMid slice" clip-path="url(#avatar)" />
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" class="ring" stroke-width="1" />
  <text x="%.1f" y="%d" text-anchor="middle" font-size="11" font-weight="500" font-family="%s" class="name">%s</text>
  <text x="%.1f" y="%d" text-anchor="middle" font-size="11" letter-spacing="0.5" font-family="%s" class="muted">%s</text>
</svg>
`,
		contributorsCellWidth, contributorsCellHeight, contributorsCellWidth, contributorsCellHeight,
		html.EscapeString(contributor.Name),
		avatar, center-float64(contributorsAvatar)/2, contributorsAvatar, contributorsAvatar,
		center, float64(contributorsAvatar)/2, float64(contributorsAvatar)/2+0.5,
		center, contributorsAvatar+17, contributorsFont, html.EscapeString(truncateName(contributor.Name)),
		center, contributorsAvatar+35, contributorsEmojiFont, html.EscapeString(contributorEmoji(contributor.Contributions)))
}

// getContributorAvatarCell draws a contributor as an avatar alone, for the roll the documentation
// overlords are given. It is the same drawing as a full cell with the name and the emoji left off,
// as both say the same thing for everybody in that roll.
func getContributorAvatarCell(contributor Contributor, avatar string) string {
	center := float64(contributorsOverlordAvatar) / 2

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">
  <style>
    svg { color-scheme: light dark }
    .ring { stroke: rgba(24, 24, 27, 0.12) }
    @media (prefers-color-scheme: dark) {
      .ring { stroke: rgba(250, 250, 250, 0.14) }
    }
  </style>
  <defs>
    <clipPath id="avatar" clipPathUnits="objectBoundingBox"><circle cx="0.5" cy="0.5" r="0.5" /></clipPath>
  </defs>
  <image href="data:image/jpeg;base64,%s" x="0" y="0" width="%d" height="%d" preserveAspectRatio="xMidYMid slice" clip-path="url(#avatar)" />
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" class="ring" stroke-width="1" />
</svg>
`,
		contributorsOverlordAvatar, contributorsOverlordAvatar, contributorsOverlordAvatar, contributorsOverlordAvatar,
		html.EscapeString(contributor.Name),
		avatar, contributorsOverlordAvatar, contributorsOverlordAvatar,
		center, center, center-0.5)
}

// isDocumentationOverlord reports whether the documentation is the only thing a contributor is
// credited with. Two thirds of the roster is in that position, and giving every one of them a name
// and an emoji which only ever reads the same buries the rest of the card beneath them.
func isDocumentationOverlord(contributor Contributor) bool {
	for _, contribution := range contributor.Contributions {
		if contribution != contributorsContributionDoc {
			return false
		}
	}

	return len(contributor.Contributions) != 0
}

// getContributorsCard renders the roster as the subset of HTML which survives GitHub's markdown
// sanitizer: a flow of anchors, each wrapping the cell drawn for one contributor. A table would give
// the grid a border and a striped background on every second row, neither of which can be turned off
// once the sanitizer has removed the style and class attributes, so the cells are left to wrap on
// their own instead of being placed in columns.
//
// A cell is given a width but deliberately no height. GitHub reserves space for an image whose
// dimensions it knows by styling it with a muted background and rounded corners until it loads, which
// a cell would wear as a grey box for as long as the placeholder is painted.
func getContributorsCard(contributors []Contributor) string {
	buf := &strings.Builder{}

	var overlords []Contributor

	fmt.Fprintf(buf, `<p align="center"><sub>Thanks goes to these <b>%d</b> wonderful people (<a href="%s">emoji key</a>)</sub></p>
<p align="center">
`, len(contributors), contributorsEmojiKey)

	for _, contributor := range contributors {
		if isDocumentationOverlord(contributor) {
			overlords = append(overlords, contributor)

			continue
		}

		writeContributorAnchor(buf, contributor, contributorsCellWidth)
	}

	fmt.Fprint(buf, "</p>\n")

	if len(overlords) == 0 {
		return buf.String()
	}

	fmt.Fprintf(buf, `<p align="center"><b>%s Documentation Overlords</b><br><sub>and these <b>%d</b> wonderful people who keep the documentation worth reading</sub></p>
<p align="center">
`, contributorsEmoji[contributorsContributionDoc], len(overlords))

	for _, contributor := range overlords {
		writeContributorAnchor(buf, contributor, contributorsOverlordAvatar)
	}

	fmt.Fprint(buf, "</p>\n")

	return buf.String()
}

// writeContributorAnchor writes the link the README carries for a contributor, around the cell drawn
// for them.
func writeContributorAnchor(buf *strings.Builder, contributor Contributor, width int) {
	fmt.Fprintf(buf, `<a href="%s" title="%s"><img src="%s/%s%s" width="%d" alt="%s"></a>
`,
		html.EscapeString(contributor.Profile), html.EscapeString(contributor.Name),
		contributorsImageURL, contributor.ID, extSVG,
		width, html.EscapeString(contributor.Name))
}

// writeContributorsCard replaces the block between the markers, leaving the rest of the README as it
// was found.
func writeContributorsCard(path, card string) (err error) {
	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		return err
	}

	content := string(data)

	start, end := strings.Index(content, contributorsMarkerStart), strings.Index(content, contributorsMarkerEnd)

	if start == -1 || end == -1 || end < start {
		return fmt.Errorf("file '%s' must contain the '%s' and '%s' markers in that order", path, contributorsMarkerStart, contributorsMarkerEnd)
	}

	content = content[:start+len(contributorsMarkerStart)] + "\n" + card + content[end:]

	return os.WriteFile(path, []byte(content), 0600) //nolint:gosec // The path is the README of the root the generator was pointed at, which is a developer supplied flag rather than user input.
}

// contributorEmoji renders the all-contributors emoji for a contributor, collapsing a long list into
// a count so a cell cannot grow wide enough to overflow the one beside it.
func contributorEmoji(contributions []string) string {
	var emoji []string

	for _, contribution := range contributions {
		if symbol, ok := contributorsEmoji[contribution]; ok {
			emoji = append(emoji, symbol)
		}
	}

	if len(emoji) <= contributorsMaxEmoji {
		return strings.Join(emoji, "")
	}

	return fmt.Sprintf("%s+%d", strings.Join(emoji[:contributorsMaxEmoji], ""), len(emoji)-contributorsMaxEmoji)
}

func truncateName(name string) string {
	runes := []rune(name)

	if len(runes) <= contributorsMaxName {
		return name
	}

	return strings.TrimRight(string(runes[:contributorsMaxName-1]), " ") + "…"
}

// average returns the mean of a summed color channel, clamped to the channel's range.
func average(sum, n uint32) uint8 {
	if value := sum / n; value < math.MaxUint8 {
		return uint8(value)
	}

	return math.MaxUint8
}
