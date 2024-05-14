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

	"regexp"

	"github.com/Johnnycyan/elevenlabs/client"
	"github.com/Johnnycyan/elevenlabs/client/types"
	"github.com/gorilla/websocket"
	"github.com/sandisuryadi36/number-to-words/convert"
)

var (
	voices         []Voice
	voice          string
	defaultVoice   string
	defaultVoiceID string
	elevenKey      string
	ttsKey         string
	playing        = map[string]bool{}
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
		defaultVoiceID = voices[0].ID
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

func validVoice(voice string) bool {
	if voice == "" {
		return false
	}
	for _, v := range voices {
		if strings.ToLower(v.Name) == strings.ToLower(voice) {
			return true
		}
	}
	return false
}

func getVoiceID(voice string) (string, error) {
	for _, v := range voices {
		if strings.ToLower(v.Name) == strings.ToLower(voice) {
			return v.ID, nil
		}
	}
	return "", fmt.Errorf("Voice not found")
}

func configureVoice(fallbackVoice string, text string) (bool, string) {
	// set the fallback voice to the default voice if no fallback voice is provided
	if fallbackVoice == "" {
		fallbackVoice = defaultVoice
	}

	// check if the text starts with a voice name in brackets
	var customVoice string
	if strings.HasPrefix(text, "[") {
		// find the voice name in between the brackets and then remove it from the text
		voiceStart := strings.Index(text, "[")
		voiceEnd := strings.Index(text, "]")
		if voiceStart != -1 && voiceEnd != -1 {
			customVoice = strings.ToLower(text[voiceStart+1 : voiceEnd])
			text = strings.TrimSpace(text[voiceEnd+1:])
		}
	} else {
		customVoice = ""
	}

	if customVoice != "" {
		logger("Voice found in message", logDebug)
	}

	var selectedVoice string
	// check if the custom message voice is valid
	if !validVoice(customVoice) {
		if customVoice != "" {
			logger("Invalid custom message voice: "+customVoice, logDebug)
		}
		// If the custom message voice is not valid then check if the fallback voice is valid
		if !validVoice(fallbackVoice) {
			logger("Invalid fallback voice: "+fallbackVoice+" so defaulting to "+defaultVoice, logInfo)
			// If the fallback voice is not valid then default to the default voice
			selectedVoice = defaultVoiceID
		} else {
			// If the fallback voice is valid then set the selected voice to the fallback voice
			logger("Voice selected: "+fallbackVoice, logDebug)
			var err error
			selectedVoice, err = getVoiceID(fallbackVoice)
			if err != nil {
				logger("Error getting voice ID: "+err.Error(), logError)
				return false, text
			}
		}
	} else {
		// If the custom message voice is valid then set the selected voice to the custom message voice
		logger("Voice selected: "+customVoice, logDebug)
		var err error
		selectedVoice, err = getVoiceID(customVoice)
		if err != nil {
			logger("Error getting voice ID: "+err.Error(), logError)
			return false, text
		}
	}

	if selectedVoice == "" {
		logger("Invalid voice so defaulting to "+defaultVoice, logInfo)
		selectedVoice = defaultVoiceID
	}

	voice = selectedVoice

	return true, text
}

func convertNumberToWords(text string) string {
	// find numbers in the string, convert them to words and replace them in the string. I want to find them even if it's for example xdd34624 so not only numbers that are separated by spaces
	// I'm using a regex to find the numbers and then convert them to words

	// find all numbers in the string
	re := regexp.MustCompile(`\d+(\.\d+)?`) // example: 123, 123.48, xdd33444 -> 123, 123.48, 33444
	numbers := re.FindAllString(text, -1)

	if numbers == nil {
		return text
	}

	// add a space between any string character and number
	text = re.ReplaceAllString(text, " $0") // example: 123, 123.48, xdd33444 -> 123, 123.48, xdd 33444

	// convert the numbers to words
	for _, number := range numbers {
		// convert the number to words
		words := convert.NumberToWords(number, "en")

		// replace the number in the string with the words
		text = strings.Replace(text, number, words, -1)
	}

	text = strings.TrimSpace(text)
	text = strings.Replace(text, "  ", " ", -1)

	return text
}

func handleTTS(w http.ResponseWriter, r *http.Request) {
	logger("Received TTS request", logInfo)

	channel := strings.ToLower(r.URL.Query().Get("channel"))
	if channel == "" {
		http.Error(w, "channel query parameter is required", http.StatusBadRequest)
		return
	}

	if playing[channel] {
		logger("Last audio is still playing on "+channel, logInfo)
		http.Error(w, "Wait for the last TTS to finish playing", http.StatusTooManyRequests)
		return
	}
	playing[channel] = true

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
	fallbackVoice := strings.ToLower(r.URL.Query().Get("voice"))
	valid, text := configureVoice(fallbackVoice, text)
	if !valid {
		http.Error(w, "No valid voice found in URL or message", http.StatusBadRequest)
		return
	}

	text = convertNumberToWords(text)

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
