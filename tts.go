package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Johnnycyan/elevenlabs/client"
	"github.com/Johnnycyan/elevenlabs/client/types"
	"github.com/gorilla/websocket"
)

var (
	requestTime  = map[string]time.Time{}
	voices       []Voice
	voice        string
	defaultVoice string
	elevenKey    string
	ttsKey       string
)

type Voice struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func setupVoices() {
	voicesEnv := os.Getenv("VOICES")
	err := json.Unmarshal([]byte(voicesEnv), &voices)
	if err != nil {
		logger("Error unmarshalling voices.json: "+err.Error(), logError)
		return
	}
	if len(voices) > 0 {
		defaultVoice = voices[0].Name
		logger("Default voice: "+defaultVoice, logDebug)
	}
}

func handleTTSAudio(w http.ResponseWriter, _ *http.Request, text string, channel string, alert bool, stability float64, similarityBoost float64, style float64) {
	audioData, err := generateAudio(text, stability, similarityBoost, style)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if alert {
		alertSound, alertExists := getAlertSound(channel)

		if alertExists {
			alertSoundBytes, err := io.ReadAll(alertSound)
			if err != nil {
				logger("Error reading alert sound: "+err.Error(), logError)
			} else {
				for client, clientChannel := range clients {
					clientName := getClientName(fmt.Sprintf("%p", client))
					if clientChannel == channel {
						err := client.WriteMessage(websocket.BinaryMessage, alertSoundBytes)
						if err != nil {
							logger("Error sending alert sound to "+clientName+": "+err.Error(), logError)
							client.Close()
							delete(clients, client)
						} else {
							logger("Alert sound sent to "+clientName+" on channel "+channel, logInfo)
						}
					}
				}
				time.Sleep(3 * time.Second)
			}
		}
	}

	for client, clientChannel := range clients {
		if clientChannel == channel {
			clientName := getClientName(fmt.Sprintf("%p", client))
			err := client.WriteMessage(websocket.BinaryMessage, audioData)
			if err != nil {
				logger("Error sending audio data to "+clientName+": "+err.Error(), logError)
				client.Close()
				delete(clients, client)
			}
			logger("Audio data sent to "+clientName+" on channel "+channel, logInfo)
		}
	}
}

func handleTTS(w http.ResponseWriter, r *http.Request) {
	logger("Received TTS request", logInfo)

	authKey := r.URL.Query().Get("key")
	if authKey != ttsKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "No text provided.", http.StatusBadRequest)
		return
	}
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	if channel == "" {
		http.Error(w, "channel query parameter is required", http.StatusBadRequest)
		return
	}
	voice = strings.ToLower(r.URL.Query().Get("voice"))
	if voice == "" {
		voice = defaultVoice
	}
	if strings.HasPrefix(text, "[") {
		// find the voice name in between the brackets and then remove it from the text
		voiceStart := strings.Index(text, "[")
		voiceEnd := strings.Index(text, "]")
		if voiceStart != -1 && voiceEnd != -1 {
			voice = strings.ToLower(text[voiceStart+1 : voiceEnd])
			text = strings.TrimSpace(text[voiceEnd+1:])
		}
	}
	stabilityString := r.URL.Query().Get("stability")
	similarityBoostString := r.URL.Query().Get("similarityBoost")
	styleString := r.URL.Query().Get("style")

	if stabilityString == "" {
		stabilityString = "0.40"
	}
	if similarityBoostString == "" {
		similarityBoostString = "1.00"
	}
	if styleString == "" {
		styleString = "0.00"
	}

	stability, err := strconv.ParseFloat(stabilityString, 64)
	if err != nil {
		logger("Invalid stability: "+stabilityString+" so defaulting to 0.40", logInfo)
		stability = 0.40
	}
	similarityBoost, err := strconv.ParseFloat(similarityBoostString, 64)
	if err != nil {
		logger("Invalid similarityBoost: "+similarityBoostString+" so defaulting to 1.00", logInfo)
		similarityBoost = 1.00
	}
	style, err := strconv.ParseFloat(styleString, 64)
	if err != nil {
		logger("Invalid style: "+styleString+" so defaulting to 0.00", logInfo)
		style = 0.00
	}
	// check if the voice is valid
	var selectedVoice string
	for _, v := range voices {
		if strings.ToLower(v.Name) == strings.ToLower(voice) {
			selectedVoice = v.ID
			break
		}
	}

	if selectedVoice == "" {
		logger("Invalid voice: "+voice+" so defaulting to the first voice.", logInfo)
		selectedVoice = voices[0].ID
	} else {
		logger("Voice selected: "+voice, logDebug)
	}

	voice = selectedVoice

	if len(clients) == 0 {
		logger("No connected clients", logInfo)
		http.Error(w, "No connected clients", http.StatusNotFound)
		return
	}

	// Check if there is a connected client for the channel
	found := false
	for client, clientChannel := range clients {
		clientName := getClientName(fmt.Sprintf("%p", client))
		if clientChannel == channel {
			logger("Found client "+clientName+" for channel "+channel, logDebug)
			found = true
			break
		}
	}
	if found == false {
		logger("No connected client for channel "+channel, logInfo)
		http.Error(w, "No connected client for channel", http.StatusNotFound)
		return
	}

	if time.Since(requestTime[channel]) < 10*time.Second {
		logger("Rate limit exceeded for channel "+channel, logInfo)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	requestTime[channel] = time.Now()

	go handleTTSAudio(w, r, text, channel, false, stability, similarityBoost, style)
	return
}

func generateAudio(text string, stability float64, similarityBoost float64, style float64) ([]byte, error) {
	logger("Generating audio for text: "+text, logDebug)

	ctx := context.Background()
	client := client.New(elevenKey)
	pipeReader, pipeWriter := io.Pipe()

	clientData, err := client.GetUserInfo(ctx)
	if err != nil {
		return nil, err
	}

	userTier := strings.TrimSpace(clientData.Subscription.Tier)
	var format string
	if userTier == "starter" {
		format = "mp3_44100_128"
	} else if userTier == "creator" {
		format = "mp3_44100_192"
	} else {
		format = "mp3_44100_128"
	}

	go func() {
		err := client.TTSStream(ctx, pipeWriter, text, "eleven_multilingual_v2", voice, types.SynthesisOptions{Stability: stability, SimilarityBoost: similarityBoost, Format: format, Style: style})
		if err != nil {
			logger(err.Error(), logError)
		}
		pipeWriter.Close()
	}()

	audioData, err := io.ReadAll(pipeReader)
	if err != nil {
		return nil, err
	}

	return audioData, nil
}
