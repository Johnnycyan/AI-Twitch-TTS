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

	"github.com/Johnnycyan/elevenlabs/client"
	"github.com/Johnnycyan/elevenlabs/client/types"
)

var (
	voices         []Voice
	voiceModels    []VoiceModel
	voiceStyles    []VoiceStyle
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
		logger("Error unmarshalling voices.json: "+err.Error(), logError, "Universal")
		return
	}
	if len(voices) > 0 {
		defaultVoice = voices[0].Name
		defaultVoiceID = voices[0].ID
		logger("Default voice: "+defaultVoice, logDebug, "Universal")
	}
}

func setupVoiceModels() {
	voiceModelsEnv := os.Getenv("VOICE_MODELS")
	err := json.Unmarshal([]byte(voiceModelsEnv), &voiceModels)
	if err != nil {
		logger("Error unmarshalling voice models: "+err.Error(), logError, "Universal")
		return
	}
}

func setupVoiceStyles() {
	voiceStylesEnv := os.Getenv("VOICE_STYLES")
	err := json.Unmarshal([]byte(voiceStylesEnv), &voiceStyles)
	if err != nil {
		logger("Error unmarshalling voice styles: "+err.Error(), logError, "Universal")
		return
	}
}

func validVoice(voice string) bool {
	if voice == "" {
		return false
	}
	for _, v := range voices {
		if strings.EqualFold(v.Name, voice) {
			return true
		}
	}
	return false
}

func getVoiceID(voice string) (string, error) {
	for _, v := range voices {
		if strings.EqualFold(v.Name, voice) {
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

func getVoiceModel(ID string) (string, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError, "Universal")
		return "", err
	}
	logger("Getting voice model for voice: "+voice, logDebug, "Universal")
	for _, v := range voiceModels {
		if strings.EqualFold(v.Name, voice) {
			return v.Model, nil
		}
	}
	logger("Voice model not found", logDebug, "Universal")
	return "", fmt.Errorf("Voice model not found")
}

func getVoiceStyle(ID string) (float64, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError, "Universal")
		return 0, err
	}
	logger("Getting voice style for voice: "+voice, logDebug, "Universal")
	for _, v := range voiceStyles {
		if strings.EqualFold(v.Name, voice) {
			style, err := strconv.ParseFloat(v.Style, 64)
			if err != nil {
				logger("Error parsing voice style: "+err.Error(), logError, "Universal")
				return 0, err
			}
			return style, nil
		}
	}
	logger("Voice style not found", logDebug, "Universal")
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

	logger("Generating TTS audio for text: "+request.Text, logDebug, request.Channel)

	voiceModifierList, err := getVoiceModifiers(request.Voice.Voice)
	if err != nil {
		logger("No voice modifiers found", logDebug, request.Channel)
	} else {
		// split the voiceModifierList into a list of voice modifiers by splitting on the comma
		voiceModifiers := strings.Split(voiceModifierList, ",")
		for _, modifier := range voiceModifiers {
			if modifier == "reverb" {
				verb = true
			}
		}
	}

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

	logger("Using model: "+model, logDebug, request.Channel)

	clientData, err := ttsClient.GetUserInfo(ctx)
	if err != nil {
		logger("Error getting user info: "+err.Error(), logError, request.Channel)
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

	logger("Using style: "+fmt.Sprintf("%f", style), logDebug, request.Channel)

	go func() {
		err := ttsClient.TTSStream(ctx, pipeWriter, request.Text, model, request.Voice.Voice, types.SynthesisOptions{Stability: request.Voice.Stability, SimilarityBoost: request.Voice.SimilarityBoost, Format: format, Style: style})
		if err != nil {
			logger("Error generating TTS audio: "+err.Error(), logError, request.Channel)
		}
		pipeWriter.Close()
	}()

	audioData, err := io.ReadAll(pipeReader)
	if err != nil {
		logger("Error reading TTS audio data: "+err.Error(), logError, request.Channel)
		return nil, err
	}

	if verb {
		verbAudio := reverb(audioData, request.Channel)
		if verbAudio == nil {
			logger("Error applying reverb to audio", logError, request.Channel)
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
		logger("Error getting user info: "+err.Error(), logError, "Universal")
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
		logger("Error marshalling client data: "+err.Error(), logError, "Universal")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(clientDataJSON)
}
