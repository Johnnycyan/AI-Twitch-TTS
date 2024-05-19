package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	serverURL string
	sentryURL string
)

func setupENV() {
	err := godotenv.Load()
	if err != nil {
		logger("Error loading .env file", logError)
	}
	elevenKey = os.Getenv("ELEVENLABS_KEY")
	serverURL = os.Getenv("SERVER_URL")
	sentryURL = os.Getenv("SENTRY_URL")
	ttsKey = os.Getenv("TTS_KEY")
	if elevenKey == "" || serverURL == "" || ttsKey == "" {
		logger("Missing required environment variables", logError)
		return
	}
	args := os.Args
	if len(args) == 2 {
		port = args[1]
	} else if len(args) == 3 {
		port = args[1]
		logLevel = args[2]
	} else if len(args) > 3 {
		log.Fatal("Too many arguments provided")
	} else {
		log.Fatal("Not enough arguments provided. Please provide at least a port number and optionally a log level.")
	}
	setupPally()
	setupVoices()
}
