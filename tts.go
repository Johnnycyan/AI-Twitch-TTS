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
	voices         []Voice
	voiceModels    []VoiceModel
	voiceStyles    []VoiceStyle
	voice          string
	defaultVoice   string
	defaultVoiceID string
	elevenKey      string
	ttsClient      client.Client
	ttsKey         string
)

type Voice struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type VoiceModel struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

type VoiceStyle struct {
	Name  string `json:"name"`
	Style string `json:"style"`
}

type TTSSettings struct {
	Voice           string
	Stability       float64
	SimilarityBoost float64
	Style           float64
}

func createClient() {
	ttsClient = client.New(elevenKey)
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

func setupVoiceModels() {
	voiceModelsEnv := os.Getenv("VOICE_MODELS")
	err := json.Unmarshal([]byte(voiceModelsEnv), &voiceModels)
	if err != nil {
		logger("Error unmarshalling voice models: "+err.Error(), logError)
		return
	}
}

func setupVoiceStyles() {
	voiceStylesEnv := os.Getenv("VOICE_STYLES")
	err := json.Unmarshal([]byte(voiceStylesEnv), &voiceStyles)
	if err != nil {
		logger("Error unmarshalling voice styles: "+err.Error(), logError)
		return
	}
}

func handleTTSAudio(w http.ResponseWriter, _ *http.Request, request Request, alert bool) {
	audioData, err := generateAudio(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if alert {
		alertSound, alertExists := getAlertSound(request.Channel)

		if alertExists {
			alertSoundBytes, err := io.ReadAll(alertSound)
			if err != nil {
				logger("Error reading alert sound: "+err.Error(), logError)
			} else {
				for client, clientChannel := range clients {
					clientName := getClientName(fmt.Sprintf("%p", client))
					if clientChannel == request.Channel {
						err := client.WriteMessage(websocket.BinaryMessage, alertSoundBytes)
						if err != nil {
							logger("Error sending alert sound to "+clientName+": "+err.Error(), logError)
							client.Close()
							delete(clients, client)
						} else {
							logger("Alert sound sent to "+clientName+" on channel "+request.Channel, logInfo)
						}
					}
				}
				time.Sleep(3 * time.Second)
			}
		}
	}

	sendAudio(request, audioData)

	clearChannelRequests(request.Channel)
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

func getVoiceName(ID string) (string, error) {
	for _, v := range voices {
		if v.ID == ID {
			return v.Name, nil
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

func getVoiceModel(ID string) (string, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError)
		return "", err
	}
	logger("Getting voice model for voice: "+voice, logDebug)
	for _, v := range voiceModels {
		if strings.ToLower(v.Name) == strings.ToLower(voice) {
			return v.Model, nil
		}
	}
	logger("Voice model not found", logDebug)
	return "", fmt.Errorf("Voice model not found")
}

func getVoiceStyle(ID string) (float64, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError)
		return 0, err
	}
	logger("Getting voice style for voice: "+voice, logDebug)
	for _, v := range voiceStyles {
		if strings.ToLower(v.Name) == strings.ToLower(voice) {
			style, err := strconv.ParseFloat(v.Style, 64)
			if err != nil {
				logger("Error parsing voice style: "+err.Error(), logError)
				return 0, err
			}
			return style, nil
		}
	}
	logger("Voice style not found", logDebug)
	return 0, fmt.Errorf("Voice style not found")
}

func generateAudio(request Request) ([]byte, error) {
	var verb bool
	if strings.HasPrefix(request.Text, "(reverb) ") {
		verb = true
		request.Text = strings.TrimPrefix(request.Text, "(reverb) ")
	} else {
		verb = false
	}

	logger("Generating TTS audio for text: "+request.Text, logDebug)

	ctx := context.Background()
	pipeReader, pipeWriter := io.Pipe()

	var model string
	voiceModel, err := getVoiceModel(request.Voice.Voice)
	if err != nil {
		model = "eleven_multilingual_v2"
	}

	if voiceModel != "" {
		if voiceModel == "turbo" {
			model = "eleven_turbo_v2"
		} else {
			model = "eleven_multilingual_v2"
		}
	} else {
		model = "eleven_multilingual_v2"
	}

	logger("Using model: "+model, logDebug)

	clientData, err := ttsClient.GetUserInfo(ctx)
	if err != nil {
		logger("Error getting user info: "+err.Error(), logError)
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

	var style float64
	style, err = getVoiceStyle(request.Voice.Voice)
	if err != nil {
		style = request.Voice.Style
	}

	logger("Using style: "+fmt.Sprintf("%f", style), logDebug)

	go func() {
		err := ttsClient.TTSStream(ctx, pipeWriter, request.Text, model, request.Voice.Voice, types.SynthesisOptions{Stability: request.Voice.Stability, SimilarityBoost: request.Voice.SimilarityBoost, Format: format, Style: style})
		if err != nil {
			logger("Error generating TTS audio: "+err.Error(), logError)
		}
		pipeWriter.Close()
	}()

	audioData, err := io.ReadAll(pipeReader)
	if err != nil {
		logger("Error reading TTS audio data: "+err.Error(), logError)
		return nil, err
	}

	if verb {
		verbAudio := reverb(audioData, request.Channel)
		if verbAudio == nil {
			logger("Error applying reverb to audio", logError)
			return nil, fmt.Errorf("Error applying reverb to audio")
		}
		audioData = verbAudio
	}

	return audioData, nil
}

type ClientData struct {
	CharactersLeft  int `json:"characters_left"`
	CharactersReset int `json:"characters_reset"`
}

func getCharactersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	clientInfo, err := ttsClient.GetUserInfo(ctx)
	if err != nil {
		logger("Error getting user info: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	characters := clientInfo.Subscription.CharacterCount
	characterLimit := clientInfo.Subscription.CharacterLimit

	charactersRemaining := characterLimit - characters

	charactersReset := clientInfo.Subscription.NextCharacterCountResetUnix

	clientData := ClientData{
		CharactersLeft:  int(charactersRemaining),
		CharactersReset: int(charactersReset),
	}

	clientDataJSON, err := json.Marshal(clientData)
	if err != nil {
		logger("Error marshalling client data: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(clientDataJSON)
}
