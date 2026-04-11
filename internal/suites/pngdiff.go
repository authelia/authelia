package suites

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// PNGPixelDiffResult summarizes a pixel-by-pixel comparison of two PNG byte slices.
type PNGPixelDiffResult struct {
	BaselineBounds  image.Rectangle
	ActualBounds    image.Rectangle
	TotalPixels     int
	DifferingPixels int
	Percentage      float64
}

// ComparePNGPixels decodes two PNG byte slices and counts pixels where any of the four RGBA
// channels differ. The returned Percentage is in the [0, 100] range.
func ComparePNGPixels(a, b []byte) (PNGPixelDiffResult, error) {
	var result PNGPixelDiffResult

	imgA, err := png.Decode(bytes.NewReader(a))
	if err != nil {
		return result, fmt.Errorf("decode baseline PNG: %w", err)
	}

	imgB, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return result, fmt.Errorf("decode actual PNG: %w", err)
	}

	result.BaselineBounds = imgA.Bounds()
	result.ActualBounds = imgB.Bounds()

	if result.BaselineBounds != result.ActualBounds {
		return result, nil
	}

	bnds := result.BaselineBounds
	result.TotalPixels = bnds.Dx() * bnds.Dy()

	for y := bnds.Min.Y; y < bnds.Max.Y; y++ {
		for x := bnds.Min.X; x < bnds.Max.X; x++ {
			ar, ag, ab, aa := imgA.At(x, y).RGBA()
			br, bg, bb, ba := imgB.At(x, y).RGBA()

			if ar != br || ag != bg || ab != bb || aa != ba {
				result.DifferingPixels++
			}
		}
	}

	if result.TotalPixels > 0 {
		result.Percentage = 100 * float64(result.DifferingPixels) / float64(result.TotalPixels)
	}

	return result, nil
}

// AssertVisualSnapshot compares a screenshot against a committed baseline at
// internal/suites/testdata/<name>. When UPDATE_SNAPSHOTS=1 the baseline is (re)written and
// the test passes — this is the only path that creates a baseline. A missing baseline
// without UPDATE_SNAPSHOTS=1 fails the test loudly so CI cannot accidentally green-light
// a deleted or never-committed snapshot. On byte mismatch with zero pixel diff the test
// passes (PNG encoder non-determinism). Any real pixel difference fails the test and
// writes the current screenshot to <baseline>.actual.png for local inspection.
func AssertVisualSnapshot(t *testing.T, repoRoot, name string, screenshot []byte) {
	t.Helper()

	baselinePath := filepath.Join(repoRoot, "internal", "suites", "testdata", name)

	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(baselinePath), 0755))
		require.NoError(t, os.WriteFile(baselinePath, screenshot, 0600))
		t.Logf("snapshot baseline updated at %s", baselinePath)

		return
	}

	baseline, err := os.ReadFile(baselinePath)
	if os.IsNotExist(err) {
		t.Fatalf("snapshot baseline %s does not exist; re-run with --update-snapshots to create it", baselinePath)
	}

	require.NoError(t, err)

	if bytes.Equal(baseline, screenshot) {
		return
	}

	actualPath := baselinePath + ".actual.png"
	_ = os.WriteFile(actualPath, screenshot, 0600)

	diff, pixelErr := ComparePNGPixels(baseline, screenshot)
	if pixelErr != nil {
		t.Fatalf("snapshot %s decode failed: %v (new snapshot at %s)", baselinePath, pixelErr, actualPath)
	}

	if diff.BaselineBounds != diff.ActualBounds {
		t.Fatalf("snapshot %s dimensions differ: baseline=%v actual=%v (new snapshot at %s)",
			baselinePath, diff.BaselineBounds, diff.ActualBounds, actualPath)
	}

	if diff.DifferingPixels == 0 {
		_ = os.Remove(actualPath)
		return
	}

	t.Fatalf("snapshot %s differs: %d/%d pixels (%.4f%%) (new snapshot at %s); re-run with --update-snapshots to refresh the baseline",
		baselinePath, diff.DifferingPixels, diff.TotalPixels, diff.Percentage, actualPath)
}
