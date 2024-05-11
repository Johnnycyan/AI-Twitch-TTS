package main

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	serverURL string
)

func setupENV() {
	err := godotenv.Load()
	if err != nil {
		logger("Error loading .env file", logError)
	}
	elevenKey = os.Getenv("ELEVENLABS_KEY")
	serverURL = os.Getenv("SERVER_URL")
	ttsKey = os.Getenv("TTS_KEY")
	if elevenKey == "" || serverURL == "" || ttsKey == "" {
		logger("Missing required environment variables", logError)
		return
	}
	setupPally()
	setupVoices()
}
