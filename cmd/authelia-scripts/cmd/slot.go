package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

// slotEntry is a working tree and the suite slot it owns.
type slotEntry struct {
	Slot int
	Path string
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
			fmt.Printf("%d\t%s\n", entry.Slot, entry.Path)
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
		if tree = strings.TrimSpace(string(output)); tree != "" {
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

	var fd *os.File

	if fd, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600); err != nil {
		return err
	}

	defer func() {
		_ = fd.Close()
	}()

	if err = syscall.Flock(int(fd.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("error locking slot registry '%s': %w", path, err)
	}

	defer func() {
		_ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN)
	}()

	var entries []slotEntry

	if entries, err = readSlotEntries(fd); err != nil {
		return err
	}

	if entries, err = f(entries); err != nil {
		return err
	}

	if entries == nil {
		return nil
	}

	return writeSlotEntries(fd, entries)
}

// readSlotEntries parses the registry, dropping entries whose working tree no longer exists so a deleted tree gives its
// slot back without anyone having to remember to release it.
func readSlotEntries(fd *os.File) (entries []slotEntry, err error) {
	if _, err = fd.Seek(0, 0); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		fields := strings.SplitN(scanner.Text(), "\t", 2)
		if len(fields) != 2 {
			continue
		}

		slot, errSlot := strconv.Atoi(fields[0])
		if errSlot != nil {
			continue
		}

		if _, errStat := os.Stat(fields[1]); errStat != nil {
			continue
		}

		entries = append(entries, slotEntry{Slot: slot, Path: fields[1]})
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(entries, func(a, b slotEntry) int { return a.Slot - b.Slot })

	return entries, nil
}

func writeSlotEntries(fd *os.File, entries []slotEntry) (err error) {
	builder := &strings.Builder{}

	for _, entry := range entries {
		fmt.Fprintf(builder, "%d\t%s\n", entry.Slot, entry.Path)
	}

	if err = fd.Truncate(0); err != nil {
		return err
	}

	if _, err = fd.Seek(0, 0); err != nil {
		return err
	}

	_, err = fd.WriteString(builder.String())

	return err
}

func slotEntries() (entries []slotEntry, err error) {
	err = withSlotRegistry(func(current []slotEntry) ([]slotEntry, error) {
		entries = current

		// Written back so the pruning of deleted working trees is persisted.
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
