package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	LogError LogLevel = 10 * iota
	LogWarning
	LogInfo
	LogDebug
)

func debugf(msg string, args ...interface{}) {
	if *logLevel < LogDebug {
		return
	}

	log.Printf("D "+msg, args...)
}

func infof(msg string, args ...interface{}) {
	if *logLevel < LogInfo {
		return
	}

	log.Printf("I "+msg, args...)
}

func warningf(msg string, args ...interface{}) { // nolint:deadcode
	if *logLevel < LogWarning {
		return
	}

	log.Printf("W "+msg, args...)
}

func ferrorf(w io.Writer, msg string, args ...interface{}) {
	if *logLevel < LogError {
		return
	}

	if *logLevel < LogError {
		return
	}

	// Only prefix with the level if we're logging at levels other than Error
	// or Fatal.
	if *logLevel >= LogWarning {
		fmt.Fprintf(w, "E "+msg, args...)
	}

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	fmt.Fprintf(w, msg, args...)
}

func errorf(msg string, args ...interface{}) { // nolint:deadcode
	if *logLevel < LogError {
		return
	}

	// Only prefix with the level if we're logging at levels other than Error
	// or Fatal.
	if *logLevel >= LogWarning {
		log.Printf("E "+msg, args...)
	}

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	// If we're only logging Error, we assume we're running in the context of a
	// linter, so we should strip any prefixes off the message (including the
	// date/time added by default in package log).
	fmt.Fprintf(os.Stderr, msg, args...)
}

func fatalf(msg string, args ...interface{}) {
	// Fatal messages do not get a level prefix. They end the program, so the
	// last message before a nonzero exit is always the fatal message.
	log.Fatalf(msg, args...)
}
