package suites

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	defaultDevServerStartTimeout  = 30 * time.Second
	defaultDevServerReadinessPath = "/"
)

// DevServer wraps a long-running pnpm-based dev server child process.
type DevServer struct {
	cmd      *exec.Cmd
	baseURL  string
	name     string
	stopOnce sync.Once
	stopErr  error
}

// DevServerConfig describes a pnpm-based dev server.
type DevServerConfig struct {
	Name          string
	ProjectDir    string
	Port          int
	ReadinessPath string
	StartTimeout  time.Duration
	Script        string
}

// HugoDocsDevServer is the DevServerConfig for the Hugo documentation site at docs/.
var HugoDocsDevServer = DevServerConfig{
	Name:       "hugo-docs",
	ProjectDir: "docs",
	Port:       1313,
	Script:     "ci",
}

// ReactEmailTemplatesDevServer is the DevServerConfig for the react-email template source at
// internal/templates/src.
var ReactEmailTemplatesDevServer = DevServerConfig{
	Name:       "email-templates",
	ProjectDir: "internal/templates/src",
	Port:       3000,
	Script:     "dev",
}

// StartDevServer installs the project's dependencies, spawns `pnpm dev`, and blocks until the
// server is reachable at its readiness path. onSpawn, if non-nil, is invoked as soon as the
// dev server process has been Start()-ed so callers can register it with a signal handler
// before the readiness wait below.
func StartDevServer(ctx context.Context, repoRoot string, cfg DevServerConfig, out io.Writer, onSpawn func(*DevServer)) (*DevServer, error) {
	stdoutWriter := out
	stderrWriter := out

	if stdoutWriter == nil {
		stdoutWriter = os.Stdout
	}

	if stderrWriter == nil {
		stderrWriter = os.Stderr
	}

	readinessPath := cfg.ReadinessPath
	if readinessPath == "" {
		readinessPath = defaultDevServerReadinessPath
	}

	startTimeout := cfg.StartTimeout
	if startTimeout == 0 {
		startTimeout = defaultDevServerStartTimeout
	}

	projectDir := filepath.Join(repoRoot, cfg.ProjectDir)

	install := exec.CommandContext(ctx, "pnpm", "--silent", "-C", projectDir, "install", "--frozen-lockfile")
	install.Stdout = stdoutWriter
	install.Stderr = stderrWriter

	if err := install.Run(); err != nil {
		return nil, fmt.Errorf("pnpm install for %s failed: %w", cfg.Name, err)
	}

	// Spawn the dev server in its own process group so teardown can signal the whole
	// pnpm/node/<binary> tree with a single call.
	cmd := exec.Command("pnpm", "--silent", "-C", projectDir, "run", cfg.Script)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s dev server: %w", cfg.Name, err)
	}

	srv := &DevServer{
		cmd:     cmd,
		baseURL: fmt.Sprintf("http://localhost:%d", cfg.Port),
		name:    cfg.Name,
	}

	if onSpawn != nil {
		onSpawn(srv)
	}

	if err := waitForDevServerReady(ctx, srv.baseURL+readinessPath, startTimeout); err != nil {
		_ = srv.Stop()

		return nil, fmt.Errorf("%s dev server did not become ready: %w", cfg.Name, err)
	}

	return srv, nil
}

// BaseURL returns the URL the dev server is reachable on.
func (d *DevServer) BaseURL() string {
	return d.baseURL
}

// Name returns the label the dev server was configured with.
func (d *DevServer) Name() string {
	return d.name
}

// Stop signals the dev server's process group to terminate and reaps the child.
func (d *DevServer) Stop() error {
	if d == nil || d.cmd == nil || d.cmd.Process == nil {
		return nil
	}

	d.stopOnce.Do(func() {
		if err := stopProcessGroup(d.cmd.Process.Pid); err != nil {
			d.stopErr = err

			return
		}

		_ = d.cmd.Wait()
	})

	return d.stopErr
}

func stopProcessGroup(pid int) error {
	if err := syscall.Kill(-pid, syscall.SIGINT); err != nil {
		if err != syscall.ESRCH {
			return fmt.Errorf("SIGINT to process group %d failed: %w", pid, err)
		}

		return nil
	}

	if waitForProcessGroupExit(pid, 5*time.Second) {
		return nil
	}

	_ = syscall.Kill(-pid, syscall.SIGTERM)

	if waitForProcessGroupExit(pid, 2*time.Second) {
		return nil
	}

	_ = syscall.Kill(-pid, syscall.SIGKILL)

	return nil
}

func waitForProcessGroupExit(pid int, timeout time.Duration) bool {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if syscall.Kill(-pid, 0) == syscall.ESRCH {
			return true
		}

		select {
		case <-deadline.C:
			return false
		case <-ticker.C:
		}
	}
}

func waitForDevServerReady(ctx context.Context, url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				_ = resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("dev server at %s did not become ready within %s: %w", url, timeout, ctx.Err())
		case <-ticker.C:
		}
	}
}
