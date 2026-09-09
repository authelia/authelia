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

	if root, err = cmd.Flags().GetString(cmdFlagRoot); err != nil {
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

	path := filepath.Join(pathDocsStatic, dirDocsStaticImages, fileDocsStaticImagesContributors)

	if err = os.WriteFile(path, []byte(getContributorsCard(contributors, avatars)), 0600); err != nil {
		return err
	}

	fmt.Printf("Generated %s with %d contributors.\n", path, len(contributors))

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

	for _, contributor := range config.Contributors {
		if err = validateAvatarURL(contributor.AvatarURL); err != nil {
			return nil, fmt.Errorf("error occurred validating contributor '%s': %w", contributor.Login, err)
		}

		if err = validateProfileURL(contributor.Profile); err != nil {
			return nil, fmt.Errorf("error occurred validating contributor '%s': %w", contributor.Login, err)
		}
	}

	return config.Contributors, nil
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
// supplied value which becomes an anchor in an SVG the documentation site serves, so a scheme which
// executes rather than navigates must not reach it.
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

// getContributorAvatars fetches each avatar at the density the card draws it, flattens any
// transparency onto white as JPEG has no alpha channel, and re-encodes it. Re-encoding is what keeps
// the card near a third of the size it would be with the images as GitHub serves them.
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
// the full sized original, which would otherwise be embedded at many times the density the card
// draws. An avatar already within the bound is returned untouched rather than enlarged.
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

func getContributorsCard(contributors []Contributor, avatars []string) string {
	rows := (len(contributors) + contributorsColumns - 1) / contributorsColumns
	height := contributorsGridTop + rows*contributorsRowHeight + 20
	colWidth := float64(contributorsWidth-contributorsPadding*2) / float64(contributorsColumns)

	buf := &strings.Builder{}

	fmt.Fprintf(buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="Authelia contributors">
  <style>
    svg { color-scheme: light dark }
    .rule { fill: #e4e4e7 }
    .title { fill: #18181b } .muted { fill: #71717a } .name { fill: #3f3f46 }
    .ring { stroke: rgba(24, 24, 27, 0.12) }
    @media (prefers-color-scheme: dark) {
      .rule { fill: #27272a }
      .title { fill: #fafafa } .muted { fill: #a1a1aa } .name { fill: #d4d4d8 }
      .ring { stroke: rgba(250, 250, 250, 0.14) }
    }
  </style>
  <defs>
    <clipPath id="avatar" clipPathUnits="objectBoundingBox"><circle cx="0.5" cy="0.5" r="0.5" /></clipPath>
  </defs>
  <g>
    <text x="%d" y="52" font-size="26" font-weight="700" letter-spacing="-0.02em" font-family="%s" class="title">Contributors</text>
    <text x="%d" y="76" font-size="12" font-family="%s" class="muted">Thanks goes to these %d wonderful people</text>
    <a href="%s" target="_blank" rel="noopener"><text x="%d" y="76" text-anchor="end" font-size="12" font-family="%s" text-decoration="underline" class="muted">emoji key</text></a>
    <rect x="%d" y="92" width="%d" height="1" class="rule" />
`,
		contributorsWidth, height, contributorsWidth, height,
		contributorsPadding, contributorsFont,
		contributorsPadding, contributorsFont, len(contributors),
		contributorsEmojiKey, contributorsWidth-contributorsPadding, contributorsFont,
		contributorsPadding, contributorsWidth-contributorsPadding*2)

	for i, contributor := range contributors {
		cx := float64(contributorsPadding) + colWidth*float64(i%contributorsColumns) + colWidth/2
		top := contributorsGridTop + contributorsRowHeight*(i/contributorsColumns)

		fmt.Fprintf(buf, `    <a href="%s" target="_blank" rel="noopener"><title>%s</title>
      <image href="data:image/jpeg;base64,%s" x="%.1f" y="%d" width="%d" height="%d" preserveAspectRatio="xMidYMid slice" clip-path="url(#avatar)" />
      <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" class="ring" stroke-width="1" />
      <text x="%.1f" y="%d" text-anchor="middle" font-size="11" font-weight="500" font-family="%s" class="name">%s</text>
      <text x="%.1f" y="%d" text-anchor="middle" font-size="11" letter-spacing="0.5" font-family="%s" class="muted">%s</text>
    </a>
`,
			html.EscapeString(contributor.Profile), html.EscapeString(contributor.Name),
			avatars[i], cx-float64(contributorsAvatar)/2, top, contributorsAvatar, contributorsAvatar,
			cx, float64(top)+float64(contributorsAvatar)/2, float64(contributorsAvatar)/2+0.5,
			cx, top+contributorsAvatar+17, contributorsFont, html.EscapeString(truncateName(contributor.Name)),
			cx, top+contributorsAvatar+35, contributorsEmojiFont, html.EscapeString(contributorEmoji(contributor.Contributions)))
	}

	fmt.Fprint(buf, `  </g>
</svg>
`)

	return buf.String()
}

// contributorEmoji renders the all-contributors emoji for a contributor, collapsing a long list into
// a count so a cell cannot grow wide enough to collide with the one beside it.
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
