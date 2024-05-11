package main

import (
	"log"
	"os"
)

var (
	logInfo     = "info"
	logDebug    = "debug"
	logError    = "error"
	logFountain = "fountain"
)

func logger(message string, level string) {
	args := os.Args
	var logLevel string
	if len(args) > 1 {
		logLevel = args[1]
	} else {
		logLevel = "debug"
	}

	switch level {
	case "error":
		message = "ERROR: " + message
	case "info":
		message = "INFO: " + message
	case "debug":
		message = "DEBUG: " + message
	case "fountain":
		message = "FOUNTAIN: " + message
	default:
		message = "UNKNOWN: " + message
	}

	switch logLevel {
	case "fountain":
		log.Println(message)
	case "debug":
		if level != "fountain" {
			log.Println(message)
		}
	case "info":
		if level != "debug" && level != "fountain" {
			log.Println(message)
		}
	}
}
