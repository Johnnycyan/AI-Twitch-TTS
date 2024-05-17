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

func logger(message string, level string) {
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
