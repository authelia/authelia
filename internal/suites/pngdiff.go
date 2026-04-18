package suites

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// VisualSnapshotTolerance returns the pixel-diff tolerance for the current host, adding a
// buffer on darwin since baselines are captured on Linux CI.
func VisualSnapshotTolerance(t float64) float64 {
	if runtime.GOOS == "darwin" {
		return t + 7.0
	}

	return t
}

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
// internal/suites/testdata/<name>. Pixel diffs at or below tolerancePercentage pass;
// anything above fails. Use UPDATE_SNAPSHOTS=1 to (re)write a baseline.
func AssertVisualSnapshot(t *testing.T, repoRoot, name string, screenshot []byte, tolerancePercentage float64) {
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

	if diff.Percentage <= tolerancePercentage {
		t.Logf("snapshot %s differs by %d/%d pixels (%.2f%%) — within tolerance (%.2f%%)",
			baselinePath, diff.DifferingPixels, diff.TotalPixels, diff.Percentage, tolerancePercentage)

		_ = os.Remove(actualPath)

		return
	}

	t.Fatalf("snapshot %s differs: %d/%d pixels (%.2f%%, tolerance %.2f%%) (new snapshot at %s); re-run with --update-snapshots to refresh the baseline",
		baselinePath, diff.DifferingPixels, diff.TotalPixels, diff.Percentage, tolerancePercentage, actualPath)
}
