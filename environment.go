package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var (
	serverURL    string
	sentryURL    string
	mongoEnabled bool
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
	ffmpegEnabled := strings.ToLower(os.Getenv("FFMPEG_ENABLED"))
	mongoUser = os.Getenv("MONGO_USER")
	mongoPass = os.Getenv("MONGO_PASS")
	mongoHost = os.Getenv("MONGO_HOST")
	mongoPort = os.Getenv("MONGO_PORT")
	dbName = os.Getenv("MONGO_DB")
	if elevenKey == "" || serverURL == "" || ttsKey == "" || ffmpegEnabled != "true" {
		logger("Missing required environment variables", logError)
		return
	}
	if mongoHost == "" || mongoPort == "" || dbName == "" {
		logger("MongoDB environment variables not provided. MongoDB will be disabled.", logInfo)
		mongoEnabled = false
	} else {
		logger("MongoDB environment variables provided. MongoDB will be enabled.", logInfo)
		mongoEnabled = true
	}
	createClient()
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
	setupPallyVoices()
	setupVoices()
	setupVoiceModels()
	setupVoiceStyles()
	setupVoiceModifiers()
	setupDB()
}
