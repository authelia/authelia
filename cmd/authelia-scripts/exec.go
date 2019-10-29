package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

// CommandWithStdout execute the command and forward stdout and stderr to the OS streams
func CommandWithStdout(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// RunCommandUntilCtrlC run a command until ctrl-c is hit
func RunCommandUntilCtrlC(cmd *exec.Cmd) {
	mutex := sync.Mutex{}
	cond := sync.NewCond(&mutex)
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	mutex.Lock()

	go func() {
		mutex.Lock()
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()

		fmt.Println("Hit Ctrl+C to shutdown...")

		err := cmd.Run()

		if err != nil {
			fmt.Println(err)
			cond.Broadcast()
			mutex.Unlock()
			return
		}

		<-signalChannel
		cond.Broadcast()
		mutex.Unlock()
	}()

	cond.Wait()
}
