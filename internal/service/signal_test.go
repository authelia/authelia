package service

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func TestSignalService_Run(t *testing.T) {
	testCases := []struct {
		name        string
		actionError error
		signal      os.Signal
	}{
		{
			name:        "ShouldHandleSIGHUPSuccessfully",
			actionError: nil,
			signal:      syscall.SIGHUP,
		},
		{
			name:        "ShouldHandleSIGHUPError",
			actionError: errors.New("action failed"),
			signal:      syscall.SIGHUP,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.TraceLevel)

			actionCalled := false
			action := func() (bubble bool, err error) {
				actionCalled = true
				return false, tc.actionError
			}

			service := &Signal{
				name:    "log-reload",
				signals: []os.Signal{syscall.SIGHUP},
				action:  action,
				log:     logger.WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "log-reload"}),
			}

			errChan := make(chan error, 1)
			done := make(chan struct{})

			go func() {
				err := service.Run()
				errChan <- err

				close(done)
			}()

			// Give the service a moment to start.
			time.Sleep(100 * time.Millisecond)

			p, err := os.FindProcess(os.Getpid())
			if err != nil && !errors.Is(err, tc.actionError) {
				require.NoError(t, err)
			}

			err = p.Signal(tc.signal)
			if err != nil && !errors.Is(err, tc.actionError) {
				require.NoError(t, err)
			}

			time.Sleep(100 * time.Millisecond)

			assert.NotNil(t, service.log)
			assert.NotNil(t, service.Log())

			service.Shutdown()

			select {
			case err := <-errChan:
				if err != nil && !errors.Is(err, tc.actionError) {
					require.NoError(t, err)
				}
			case <-time.After(time.Second):
				t.Fatal("service did not shut down within timeout")
			}

			assert.True(t, actionCalled, "action should have been called")
		})
	}
}

func TestSvcSignalLogReOpenFunc(t *testing.T) {
	testCases := []struct {
		name          string
		logFilePath   string
		expectService bool
	}{
		{
			name:          "ShouldCreateServiceWithLogPath",
			logFilePath:   "/var/log/authelia/authelia.log",
			expectService: true,
		},
		{
			name:          "ShouldNotCreateServiceWithoutLogPath",
			logFilePath:   "",
			expectService: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtx := newMockServiceCtx(t)
			mockCtx.config.Log.FilePath = tc.logFilePath

			service, _ := ProvisionLoggingSignal(mockCtx)

			if tc.expectService {
				require.NotNil(t, service)
				assert.Equal(t, "log-reload", service.ServiceName())
				assert.Equal(t, serviceTypeSignal, service.ServiceType())
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestLogReopenFiles(t *testing.T) {
	dir := t.TempDir()

	path := fmt.Sprintf("%s/authelia.{datetime:15:04:05.000000000}.log", dir)

	config := &schema.Configuration{
		Log: schema.Log{
			Format:     "text",
			FilePath:   path,
			KeepStdout: false,
		},
	}

	err := logging.InitializeLogger(config.Log, false)
	require.NoError(t, err)

	logging.Logger().Info("This is the first log file.")

	ctx := &testContext{
		Context:   t.Context(),
		config:    config,
		logger:    logrus.NewEntry(logging.Logger()),
		providers: middlewares.Providers{},
	}

	service, _ := ProvisionLoggingSignal(ctx)
	require.NotNil(t, service)

	errChan := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		err := service.Run()
		errChan <- err

		close(done)
	}()

	// Give the service a moment to start.
	time.Sleep(100 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)

	err = p.Signal(syscall.SIGHUP)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	logging.Logger().Info("This is the second log file.")

	service.Shutdown()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	assert.Equal(t, 2, len(entries))
}

func TestSignalService_Shutdown(t *testing.T) {
	logger := logrus.New()
	action := func() (bubble bool, err error) { return }
	service := &Signal{
		name:    "test",
		signals: []os.Signal{syscall.SIGHUP},
		action:  action,
		log:     logger.WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "test"}),
	}

	done := make(chan struct{})

	go func() {
		err := service.Run()
		assert.NoError(t, err)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	service.Shutdown()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("service did not shut down within timeout")
	}
}
