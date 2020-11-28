package main

import "log"

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

func infof(msg string, args ...interface{}) { // nolint:deadcode
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

func errorf(msg string, args ...interface{}) {
	if *logLevel < LogError {
		return
	}

	// Only prefix with the level if we're logging at levels other than Error
	// or Fatal.
	if *logLevel >= LogWarning {
		log.Printf("E "+msg, args...)
	}

	log.Printf(msg, args...)
}

func fatalf(msg string, args ...interface{}) {
	// Fatal messages do not get a level prefix. They end the program, so the
	// last message before a nonzero exit is always the fatal message.
	log.Fatalf(msg, args...)
}
