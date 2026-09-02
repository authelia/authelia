package suites

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const watchdogMargin = 5 * time.Second

var (
	watchdogMutex     sync.Mutex
	watchdogTimer     *time.Timer
	watchdogDone      chan struct{}
	watchdogCollected string
)

func startArtifactWatchdog() {
	f := flag.Lookup("test.timeout")
	if f == nil {
		return
	}

	timeout, err := time.ParseDuration(f.Value.String())
	if err != nil || timeout <= 0 {
		return
	}

	if timeout <= watchdogMargin {
		log.Debugf("Not arming the artifact watchdog: the timeout of %s leaves no room to collect before it", timeout)

		return
	}

	watchdogDone = make(chan struct{})

	watchdogTimer = time.AfterFunc(timeout-watchdogMargin, func() {
		defer close(watchdogDone)

		collectWatchdogArtifacts()
	})
}

func discardWatchdogArtifacts() {
	// Stopped so that no collection can begin once the artifacts are gone. Stop reports false once the
	// callback has been handed to a goroutine of its own, which is before it has necessarily reached the
	// lock, so the lock alone would let a discard pass an unstarted collection and leave what it went on
	// to write. Waiting for the callback to return is what closes that.
	if watchdogTimer != nil {
		if !watchdogTimer.Stop() {
			<-watchdogDone
		}

		// Cleared because Stop reports false for a timer already stopped as readily as for one that has
		// fired, so a second discard would otherwise wait on a callback that is never going to run.
		watchdogTimer = nil
	}

	watchdogMutex.Lock()
	defer watchdogMutex.Unlock()

	if watchdogCollected == "" {
		return
	}

	pattern, _ := screenshotPaths(watchdogCollected + "*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	for _, match := range matches {
		if err = os.Remove(match); err != nil {
			log.Debugf("Error discarding '%s': %v", match, err)
		}
	}

	log.Debugf("Discarded %d watchdog artifact(s): the run finished within its timeout", len(matches))

	watchdogCollected = ""
}

func collectWatchdogArtifacts() {
	watchdogMutex.Lock()
	defer watchdogMutex.Unlock()

	log.Warnf("The test binary is approaching its timeout; collecting diagnostics before it expires")

	base := "TIMEOUT"

	if suite := strings.ToUpper(os.Getenv("SUITE")); suite != "" {
		base = "TIMEOUT-" + suite
	}

	browsers := liveBrowsersSnapshot()

	// The helpers below are addressed through a session only because that is where they are defined; none
	// of them read the session itself.
	rs := &RodSession{}

	// Only the page capture creates this, and the run with no page to capture is the one whose container
	// logs are the whole of what it leaves behind: a browser that never started, or a setup that expired
	// before it could.
	if dir, _ := screenshotPaths(base); dir != "" {
		if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
			log.Debugf("Error creating the diagnostics directory: %v", err)
		}
	}

	var collected int

	for _, browser := range browsers {
		pages, err := browser.Pages()
		if err != nil {
			log.Debugf("Error listing the pages of a browser: %v", err)

			continue
		}

		for _, page := range pages {
			if err = rs.collectPage(page, fmt.Sprintf("%s-page-%d", base, collected)); err != nil {
				continue
			}

			collected++
		}
	}

	rs.collectContainerLogs(nil, base)
	rs.collectTraefikProxyAccessLog(base)

	watchdogCollected = base

	log.Warnf("Collected diagnostics for %d page(s) as '%s'", collected, base)
}
