package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestSlotShouldKeepTheSlotOfAWorkingTreeItCannotCheck(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)

	path, err := slotRegistryPath()
	require.NoError(t, err)

	unreachable := "/" + strings.Repeat("a", 300)
	require.NoError(t, os.WriteFile(path, []byte(`[{"slot":1,"path":"`+unreachable+`"}]`), 0600))

	tree := filepath.Join(t.TempDir(), "tree")
	require.NoError(t, os.Mkdir(tree, 0755))

	slot, err := allocateSlot(tree)
	require.NoError(t, err)
	assert.Equal(t, 2, slot)

	entries, err := slotEntries()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, unreachable, entries[0].Path)
}

func TestSlotShouldKeepAWorkingTreePathThatEndsInANewline(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required to resolve the working tree root")
	}

	t.Setenv("XDG_STATE_HOME", t.TempDir())

	tree := filepath.Join(t.TempDir(), "working-tree\n")
	require.NoError(t, os.Mkdir(tree, 0755))

	require.NoError(t, os.WriteFile(filepath.Join(tree, "go.mod"), []byte("module github.com/authelia/authelia/v4\n"), 0600))
	require.NoError(t, exec.Command("git", "-C", tree, "init", "-q").Run())

	t.Chdir(tree)

	resolved, err := workingTree()
	require.NoError(t, err)

	expected, err := filepath.EvalSymlinks(tree)
	require.NoError(t, err)
	require.Equal(t, expected, resolved)

	slot, err := allocateSlot(resolved)
	require.NoError(t, err)
	assert.Equal(t, 1, slot)

	again, err := allocateSlot(resolved)
	require.NoError(t, err)
	assert.Equal(t, slot, again)

	require.NoError(t, releaseSlot(resolved))

	entries, err := slotEntries()
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestSlotShouldKeepWorkingTreePathsThatContainNewlines(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	root := t.TempDir()

	odd := filepath.Join(root, "working\ntree")
	require.NoError(t, os.Mkdir(odd, 0755))

	plain := filepath.Join(root, "plain")
	require.NoError(t, os.Mkdir(plain, 0755))

	first, err := allocateSlot(odd)
	require.NoError(t, err)
	assert.Equal(t, 1, first)

	second, err := allocateSlot(plain)
	require.NoError(t, err)
	assert.Equal(t, 2, second)

	again, err := allocateSlot(odd)
	require.NoError(t, err)
	assert.Equal(t, first, again)

	entries, err := slotEntries()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, odd, entries[0].Path)
}

func TestSlotShouldRefuseARegistryThatContradictsItself(t *testing.T) {
	testCases := []struct {
		name     string
		registry func(first, second string) string
	}{
		{
			"ShouldRefuseASlotHeldTwice",
			func(first, second string) string {
				return `[{"slot":1,"path":"` + first + `"},{"slot":1,"path":"` + second + `"}]`
			},
		},
		{
			"ShouldRefuseAWorkingTreeHoldingTwoSlots",
			func(first, _ string) string {
				return `[{"slot":1,"path":"` + first + `"},{"slot":2,"path":"` + first + `"}]`
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_STATE_HOME", t.TempDir())

			root := t.TempDir()

			first, second := filepath.Join(root, "first"), filepath.Join(root, "second")
			require.NoError(t, os.Mkdir(first, 0755))
			require.NoError(t, os.Mkdir(second, 0755))

			path, err := slotRegistryPath()
			require.NoError(t, err)

			require.NoError(t, os.WriteFile(path, []byte(tc.registry(first, second)), 0600))

			_, err = slotEntries()
			assert.Error(t, err)
		})
	}
}

func TestSlotShouldRefuseARegistryItCannotRead(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)

	tree := t.TempDir()

	path, err := slotRegistryPath()
	require.NoError(t, err)

	testCases := []struct {
		name    string
		content string
	}{
		{"ShouldRefuseContentItCannotParse", "not json at all"},
		{"ShouldRefuseAnEntryWithoutAWorkingTree", `[{"slot":1,"path":""}]`},
		{"ShouldRefuseAnEntryWithoutASlot", `[{"slot":0,"path":"/tmp"}]`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0600))

			_, err = allocateSlot(tree)
			assert.Error(t, err)
		})
	}
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
