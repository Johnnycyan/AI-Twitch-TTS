package main

import (
	"log"
)

var (
	logInfo     = "info"
	logDebug    = "debug"
	logError    = "error"
	logFountain = "fountain"
	logLevel    = "debug"
)

func logger(message string, level string, channel string) {
	switch level {
	case "error":
		message = "ERROR: [" + channel + "] " + message
	case "info":
		message = "INFO: [" + channel + "] " + message
	case "debug":
		message = "DEBUG: [" + channel + "] " + message
	case "fountain":
		message = "FOUNTAIN: [" + channel + "] " + message
	default:
		message = "UNKNOWN: [" + channel + "] " + message
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
