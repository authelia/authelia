// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type stackFrame struct {
	File string
	Line int
	Name string
}

// String provides the standard file:line representation.
func (f stackFrame) String() string {
	return f.File + ":" + strconv.Itoa(f.Line) + " " + f.Name
}

type stackTrace []stackFrame

// String provides the standard multi-line stack trace with the function names aligned.
func (s stackTrace) String() string {
	var width int

	for _, frame := range s {
		if n := len(frame.File) + len(strconv.Itoa(frame.Line)) + 1; n > width {
			width = n
		}
	}

	buf := &strings.Builder{}

	for i, frame := range s {
		if i != 0 {
			buf.WriteRune('\n')
		}

		line := strconv.Itoa(frame.Line)

		buf.WriteString(frame.File)
		buf.WriteRune(':')
		buf.WriteString(line)
		buf.WriteString(strings.Repeat(" ", width-len(frame.File)-len(line)))
		buf.WriteString(frame.Name)
	}

	return buf.String()
}

type stackHook struct {
	CallerLevels []logrus.Level
	StackLevels  []logrus.Level
}

// Levels provides the levels to filter.
func (hook stackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called by logrus when something is logged.
func (hook stackHook) Fire(entry *logrus.Entry) error {
	skip := stackSkipFrames

	if len(entry.Data) != 0 {
		skip = stackSkipFramesFields
	}

	frames := stackCallers(skip)

	if len(frames) == 0 {
		return nil
	}

	if slices.Contains(hook.CallerLevels, entry.Level) {
		entry.Data[FieldCaller] = frames[0]
	}

	if slices.Contains(hook.StackLevels, entry.Level) {
		entry.Data[FieldStack] = frames
	}

	return nil
}

func stackCallers(skip int) (frames stackTrace) {
	pcs := make([]uintptr, stackMaxDepth)

	n := runtime.Callers(skip+2, pcs)

	frames = make(stackTrace, 0, n)

	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)

		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc - 1)

		if strings.Contains(file, pathPackageLogrus) {
			continue
		}

		frames = append(frames, stackFrame{File: file, Line: line, Name: stackStripPackage(fn.Name())})
	}

	return frames
}

func stackStripPackage(name string) string {
	slash := strings.LastIndex(name, "/")

	if slash == -1 {
		slash = 0
	}

	dot := strings.Index(name[slash:], ".")

	if dot == -1 {
		return name
	}

	return name[slash+dot+1:]
}
