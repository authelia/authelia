package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlotShouldAllocateTheLowestFreeNumberPerWorkingTree(t *testing.T) {
	trees := newSlotRegistry(t, 3)

	first, err := allocateSlot(trees[0])
	require.NoError(t, err)
	assert.Equal(t, 1, first)

	second, err := allocateSlot(trees[1])
	require.NoError(t, err)
	assert.Equal(t, 2, second)

	third, err := allocateSlot(trees[2])
	require.NoError(t, err)
	assert.Equal(t, 3, third)
}

func TestSlotShouldReturnTheSameNumberForAWorkingTreeThatAlreadyHasOne(t *testing.T) {
	trees := newSlotRegistry(t, 1)

	first, err := allocateSlot(trees[0])
	require.NoError(t, err)

	second, err := allocateSlot(trees[0])
	require.NoError(t, err)

	assert.Equal(t, first, second)
}

func TestSlotShouldReuseTheNumberOfAReleasedWorkingTree(t *testing.T) {
	trees := newSlotRegistry(t, 3)

	for _, tree := range trees[:2] {
		_, err := allocateSlot(tree)
		require.NoError(t, err)
	}

	require.NoError(t, releaseSlot(trees[0]))

	slot, err := allocateSlot(trees[2])
	require.NoError(t, err)
	assert.Equal(t, 1, slot)
}

func TestSlotShouldPruneWorkingTreesThatNoLongerExist(t *testing.T) {
	trees := newSlotRegistry(t, 3)

	for _, tree := range trees[:2] {
		_, err := allocateSlot(tree)
		require.NoError(t, err)
	}

	require.NoError(t, os.Remove(trees[0]))

	entries, err := slotEntries()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, trees[1], entries[0].Path)

	slot, err := allocateSlot(trees[2])
	require.NoError(t, err)
	assert.Equal(t, 1, slot)
}

// newSlotRegistry points the registry at a temporary directory and returns n working tree paths that exist on disk.
func newSlotRegistry(t *testing.T, n int) (trees []string) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	root := t.TempDir()

	trees = make([]string, n)

	for i := range trees {
		trees[i] = filepath.Join(root, string(rune('a'+i)))

		require.NoError(t, os.Mkdir(trees[i], 0755))
	}

	return trees
}
