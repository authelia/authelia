package main

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWritePromptHidden(t *testing.T) {
	var buf bytes.Buffer

	err := WritePromptHidden(&buf, "Password: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "PROMPT_HIDDEN:Password: \n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWritePromptVisible(t *testing.T) {
	var buf bytes.Buffer

	err := WritePromptVisible(&buf, "TOTP Code: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "PROMPT_VISIBLE:TOTP Code: \n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriteInfo(t *testing.T) {
	var buf bytes.Buffer

	err := WriteInfo(&buf, "Touch your security key...")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "INFO:Touch your security key...\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriteSuccess(t *testing.T) {
	var buf bytes.Buffer

	err := WriteSuccess(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "SUCCESS\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestWriteFailure(t *testing.T) {
	var buf bytes.Buffer

	err := WriteFailure(&buf, "authentication failed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "FAILURE:authentication failed\n"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestReadLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "hello\n", "hello", false},
		{"with carriage return", "hello\r\n", "hello", false},
		{"empty line", "\n", "", false},
		{"no newline EOF", "", "", true},
		{"username", "jdoe\n", "jdoe", false},
		{"password with special chars", "p@ss:w0rd!\n", "p@ss:w0rd!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))

			got, err := ReadLine(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("ReadLine() = %q, want %q", got, tt.want)
			}
		})
	}
}
