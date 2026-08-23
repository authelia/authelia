package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/utils"
)

var (
	slotList    bool
	slotRelease bool
)

type slotEntry struct {
	Slot int    `json:"slot"`
	Path string `json:"path"`
}

func newSuitesSlotCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:     "slot",
		Short:   cmdSuitesSlotShort,
		Long:    cmdSuitesSlotLong,
		Example: cmdSuitesSlotExample,
		Run:     cmdSuitesSlotRun,
		Args:    cobra.NoArgs,

		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&slotList, "list", false, "Lists every allocated slot instead of allocating one")
	cmd.Flags().BoolVar(&slotRelease, "release", false, "Releases the slot allocated to this working tree")

	return cmd
}

func cmdSuitesSlotRun(_ *cobra.Command, _ []string) {
	if slotList && slotRelease {
		log.Fatal(errors.New("only one of --list and --release may be given"))
	}

	tree, err := workingTree()
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case slotList:
		entries, err := slotEntries()
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range entries {
			fmt.Printf("%d\t%q\n", entry.Slot, entry.Path)
		}
	case slotRelease:
		if err = releaseSlot(tree); err != nil {
			log.Fatal(err)
		}
	default:
		slot, err := allocateSlot(tree)
		if err != nil {
			log.Fatal(err)
		}

		// The number alone, so that bootstrap.sh can capture it.
		fmt.Println(slot)
	}
}

// workingTree returns the root of the working tree this command runs in, resolved through any symlink. Allocation and
// release key the registry on it rather than on the current directory so that a worktree holds one slot however deep in
// it the command is run from.
func workingTree() (tree string, err error) {
	output, err := utils.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		if tree = strings.TrimSuffix(string(output), "\n"); tree != "" {
			return resolveSymlinks(tree), nil
		}
	}

	if tree, err = os.Getwd(); err != nil {
		return "", err
	}

	return resolveSymlinks(tree), nil
}

func resolveSymlinks(tree string) string {
	resolved, err := filepath.EvalSymlinks(tree)
	if err != nil {
		return tree
	}

	return resolved
}

// slotRegistryPath is the file recording which working tree owns which slot. It lives outside any working tree because
// the whole point is to keep separate working trees on one machine from picking the same number.
func slotRegistryPath() (path string, err error) {
	state := os.Getenv("XDG_STATE_HOME")

	if state == "" {
		var home string

		if home, err = os.UserHomeDir(); err != nil {
			return "", err
		}

		state = filepath.Join(home, ".local", "state")
	}

	dir := filepath.Join(state, "authelia")

	if err = os.MkdirAll(dir, 0755); err != nil { //nolint:gosec // The directory is derived from XDG_STATE_HOME, not from user input.
		return "", err
	}

	return filepath.Join(dir, "suite-slots"), nil
}

// withSlotRegistry runs f against the registry entries under an exclusive lock, writing back whatever f returns. Every
// mutation goes through here so two shells sourcing bootstrap.sh at once cannot be handed the same slot.
func withSlotRegistry(f func(entries []slotEntry) ([]slotEntry, error)) (err error) {
	var path string

	if path, err = slotRegistryPath(); err != nil {
		return err
	}

	var unlock func()

	if unlock, err = lockSlotRegistry(path); err != nil {
		return err
	}

	defer unlock()

	var entries []slotEntry

	if entries, err = readSlotEntries(path); err != nil {
		return err
	}

	if entries, err = f(entries); err != nil {
		return err
	}

	if entries == nil {
		return nil
	}

	return writeSlotEntries(path, entries)
}

// lockSlotRegistry takes the exclusive lock guarding the registry. The lock lives in a file of its own because the
// registry is replaced by a rename, and a lock held on a file that has been replaced no longer excludes anybody: the
// next process opens the new inode and locks that instead.
func lockSlotRegistry(path string) (unlock func(), err error) {
	var fd *os.File

	if fd, err = os.OpenFile(path+".lock", os.O_RDWR|os.O_CREATE, 0600); err != nil {
		return nil, err
	}

	if err = syscall.Flock(int(fd.Fd()), syscall.LOCK_EX); err != nil {
		_ = fd.Close()

		return nil, fmt.Errorf("error locking slot registry '%s': %w", path, err)
	}

	return func() {
		_ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN)
		_ = fd.Close()
	}, nil
}

// readSlotEntries parses the registry, dropping entries whose working tree no longer exists so a deleted tree gives its
// slot back without anyone having to remember to release it. A registry that cannot be read is an error rather than an
// empty one: treating it as empty would hand out numbers that other working trees are still using.
func readSlotEntries(path string) (entries []slotEntry, err error) {
	var data []byte

	if data, err = os.ReadFile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	var stored []slotEntry

	if err = json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf("error parsing slot registry '%s': %w", path, err)
	}

	slots, paths := make(map[int]bool, len(stored)), make(map[string]bool, len(stored))

	for _, entry := range stored {
		if entry.Slot < 1 || entry.Path == "" {
			return nil, fmt.Errorf("error parsing slot registry '%s': the entry for slot %d has no usable working tree", path, entry.Slot)
		}

		if _, errStat := os.Stat(entry.Path); errStat != nil {
			if errors.Is(errStat, os.ErrNotExist) {
				continue
			}

			log.Warnf("Keeping suite slot %d for '%s' because it could not be checked: %v", entry.Slot, entry.Path, errStat)
		}

		if slots[entry.Slot] {
			return nil, fmt.Errorf("error parsing slot registry '%s': slot %d is held by more than one working tree", path, entry.Slot)
		}

		if paths[entry.Path] {
			return nil, fmt.Errorf("error parsing slot registry '%s': the working tree '%s' holds more than one slot", path, entry.Path)
		}

		slots[entry.Slot], paths[entry.Path] = true, true

		entries = append(entries, entry)
	}

	slices.SortFunc(entries, func(a, b slotEntry) int { return a.Slot - b.Slot })

	return entries, nil
}

// writeSlotEntries replaces the registry by renaming a complete copy over it, so a process that dies mid-write leaves
// the previous registry rather than a truncated one. Losing it would free every slot at once.
func writeSlotEntries(path string, entries []slotEntry) (err error) {
	var data []byte

	if data, err = json.Marshal(entries); err != nil {
		return err
	}

	dir := filepath.Dir(path)

	var tmp *os.File

	if tmp, err = os.CreateTemp(dir, "suite-slots"); err != nil {
		return err
	}

	name := tmp.Name()

	defer func() {
		if err != nil {
			_ = os.Remove(name)
		}
	}()

	if _, err = tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()

		return err
	}

	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()

		return err
	}

	if err = tmp.Close(); err != nil {
		return err
	}

	if err = os.Rename(name, path); err != nil {
		return err
	}

	if errSync := syncDirectory(dir); errSync != nil {
		log.Warnf("Could not flush the slot registry directory '%s': %v", dir, errSync)
	}

	return nil
}

// syncDirectory flushes the directory entry, so the rename itself survives a crash rather than only the contents of the
// file it points at.
func syncDirectory(dir string) (err error) {
	var fd *os.File

	if fd, err = os.Open(dir); err != nil {
		return err
	}

	defer func() {
		_ = fd.Close()
	}()

	return fd.Sync()
}

func slotEntries() (entries []slotEntry, err error) {
	err = withSlotRegistry(func(current []slotEntry) ([]slotEntry, error) {
		entries = current

		return current, nil
	})

	return entries, err
}

// allocateSlot returns the slot this working tree owns, giving it the lowest unused number if it has none.
func allocateSlot(tree string) (slot int, err error) {
	err = withSlotRegistry(func(entries []slotEntry) ([]slotEntry, error) {
		for _, entry := range entries {
			if entry.Path == tree {
				slot = entry.Slot

				return entries, nil
			}
		}

		slot = 1

		for _, entry := range entries {
			if entry.Slot == slot {
				slot++
			}
		}

		return append(entries, slotEntry{Slot: slot, Path: tree}), nil
	})

	return slot, err
}

func releaseSlot(tree string) (err error) {
	return withSlotRegistry(func(entries []slotEntry) ([]slotEntry, error) {
		return slices.DeleteFunc(entries, func(entry slotEntry) bool { return entry.Path == tree }), nil
	})
}
