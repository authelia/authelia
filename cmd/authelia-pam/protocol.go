package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	cmdPromptHidden  = "PROMPT_HIDDEN:"
	cmdPromptVisible = "PROMPT_VISIBLE:"
	cmdInfo          = "INFO:"
	cmdSuccess       = "SUCCESS"
	cmdFailure       = "FAILURE:"
)

// WritePromptHidden sends a hidden-input prompt to the C shim via PAM_PROMPT_ECHO_OFF.
func WritePromptHidden(w io.Writer, text string) error {
	_, err := fmt.Fprintf(w, "%s%s\n", cmdPromptHidden, text)
	return err
}

// WritePromptVisible sends a visible-input prompt to the C shim via PAM_PROMPT_ECHO_ON.
func WritePromptVisible(w io.Writer, text string) error {
	_, err := fmt.Fprintf(w, "%s%s\n", cmdPromptVisible, text)
	return err
}

// WriteInfo sends an informational message to the C shim via PAM_TEXT_INFO.
func WriteInfo(w io.Writer, text string) error {
	_, err := fmt.Fprintf(w, "%s%s\n", cmdInfo, text)
	return err
}

// WriteSuccess signals successful authentication to the C shim.
func WriteSuccess(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s\n", cmdSuccess)
	return err
}

// WriteFailure signals failed authentication to the C shim with a reason.
func WriteFailure(w io.Writer, message string) error {
	_, err := fmt.Fprintf(w, "%s%s\n", cmdFailure, message)
	return err
}

// ReadLine reads a single line from the reader, trimming the trailing newline.
func ReadLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimRight(line, "\r\n"), nil
}
