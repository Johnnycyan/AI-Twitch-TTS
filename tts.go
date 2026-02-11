package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	voices              []Voice
	voiceModels         []VoiceModel
	voiceStyles         []VoiceStyle
	voiceSpeeds         []VoiceSpeed
	voiceSpeakerBoosts  []VoiceSpeakerBoost
	voiceLanguages      []VoiceLanguage
	voiceStabilities    []VoiceStability
	defaultVoice        string
	defaultVoiceID      string
	elevenKey           string
	ttsKey              string
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

type VoiceSpeed struct {
	Name  string `json:"name"`
	Speed string `json:"speed"`
}

type VoiceSpeakerBoost struct {
	Name         string `json:"name"`
	SpeakerBoost string `json:"speaker_boost"`
}

type VoiceLanguage struct {
	Name         string `json:"name"`
	LanguageCode string `json:"language_code"`
}

type VoiceStability struct {
	Name      string `json:"name"`
	Stability string `json:"stability"`
}

type TTSSettings struct {
	Voice           string
	Stability       float64
	SimilarityBoost float64
	Style           float64
	Speed           float64
	UseSpeakerBoost *bool
	LanguageCode    string
}

// ElevenLabs API response types
type elevenLabsSubscription struct {
	Tier                       string `json:"tier"`
	CharacterCount             int    `json:"character_count"`
	CharacterLimit             int    `json:"character_limit"`
	NextCharacterCountResetUnix int   `json:"next_character_count_reset_unix"`
}

type elevenLabsUserInfo struct {
	Subscription elevenLabsSubscription `json:"subscription"`
}

type elevenLabsVoiceInfo struct {
	PreviewURL string `json:"preview_url"`
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

func setupVoiceSpeeds() {
	voiceSpeedsEnv := os.Getenv("VOICE_SPEEDS")
	if voiceSpeedsEnv == "" {
		return
	}
	err := json.Unmarshal([]byte(voiceSpeedsEnv), &voiceSpeeds)
	if err != nil {
		logger("Error unmarshalling voice speeds: "+err.Error(), logError, "Universal")
		return
	}
}

func setupVoiceSpeakerBoosts() {
	voiceSpeakerBoostsEnv := os.Getenv("VOICE_SPEAKER_BOOSTS")
	if voiceSpeakerBoostsEnv == "" {
		return
	}
	err := json.Unmarshal([]byte(voiceSpeakerBoostsEnv), &voiceSpeakerBoosts)
	if err != nil {
		logger("Error unmarshalling voice speaker boosts: "+err.Error(), logError, "Universal")
		return
	}
}

func setupVoiceLanguages() {
	voiceLanguagesEnv := os.Getenv("VOICE_LANGUAGES")
	if voiceLanguagesEnv == "" {
		return
	}
	err := json.Unmarshal([]byte(voiceLanguagesEnv), &voiceLanguages)
	if err != nil {
		logger("Error unmarshalling voice languages: "+err.Error(), logError, "Universal")
		return
	}
}

func setupVoiceStabilities() {
	voiceStabilitiesEnv := os.Getenv("VOICE_STABILITIES")
	if voiceStabilitiesEnv == "" {
		return
	}
	err := json.Unmarshal([]byte(voiceStabilitiesEnv), &voiceStabilities)
	if err != nil {
		logger("Error unmarshalling voice stabilities: "+err.Error(), logError, "Universal")
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

func getVoiceSpeed(ID string) (float64, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError, "Universal")
		return 1.0, err
	}
	logger("Getting voice speed for voice: "+voice, logDebug, "Universal")
	for _, v := range voiceSpeeds {
		if strings.EqualFold(v.Name, voice) {
			speed, err := strconv.ParseFloat(v.Speed, 64)
			if err != nil {
				logger("Error parsing voice speed: "+err.Error(), logError, "Universal")
				return 1.0, err
			}
			return speed, nil
		}
	}
	logger("Voice speed not found", logDebug, "Universal")
	return 1.0, fmt.Errorf("Voice speed not found")
}

func getVoiceSpeakerBoost(ID string) (bool, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError, "Universal")
		return true, err
	}
	logger("Getting voice speaker boost for voice: "+voice, logDebug, "Universal")
	for _, v := range voiceSpeakerBoosts {
		if strings.EqualFold(v.Name, voice) {
			boost, err := strconv.ParseBool(v.SpeakerBoost)
			if err != nil {
				logger("Error parsing voice speaker boost: "+err.Error(), logError, "Universal")
				return true, err
			}
			return boost, nil
		}
	}
	logger("Voice speaker boost not found", logDebug, "Universal")
	return true, fmt.Errorf("Voice speaker boost not found")
}

func getVoiceLanguage(ID string) (string, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError, "Universal")
		return "", err
	}
	logger("Getting voice language for voice: "+voice, logDebug, "Universal")
	for _, v := range voiceLanguages {
		if strings.EqualFold(v.Name, voice) {
			return v.LanguageCode, nil
		}
	}
	logger("Voice language not found", logDebug, "Universal")
	return "", fmt.Errorf("Voice language not found")
}

func getVoiceStability(ID string) (float64, error) {
	voice, err := getVoiceName(ID)
	if err != nil {
		logger("Error getting voice name: "+err.Error(), logError, "Universal")
		return 0, err
	}
	logger("Getting voice stability for voice: "+voice, logDebug, "Universal")
	for _, v := range voiceStabilities {
		if strings.EqualFold(v.Name, voice) {
			stability, err := strconv.ParseFloat(v.Stability, 64)
			if err != nil {
				logger("Error parsing voice stability: "+err.Error(), logError, "Universal")
				return 0, err
			}
			return stability, nil
		}
	}
	logger("Voice stability not found", logDebug, "Universal")
	return 0, fmt.Errorf("Voice stability not found")
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
		model = "eleven_v3"
	}

	if voiceModel != "" {
		switch voiceModel {
		case "turbo":
			model = "eleven_turbo_v2"
		case "v2":
			model = "eleven_multilingual_v2"
		case "v3":
			model = "eleven_v3"
		default:
			model = "eleven_v3"
		}
	} else {
		model = "eleven_v3"
	}

	logger("Using model: "+model, logDebug, request.Channel)

	userInfo, err := getUserInfo(ctx)
	if err != nil {
		logger("Error getting user info: "+err.Error(), logError, request.Channel)
		return nil, err
	}

	userTier := strings.TrimSpace(userInfo.Subscription.Tier)
	var format string
	switch userTier {
	case "starter":
		format = "mp3_44100_128"
	case "creator":
		format = "mp3_44100_192"
	default:
		format = "mp3_44100_128"
	}

	var style float64
	style, err = getVoiceStyle(request.Voice.Voice)
	if err != nil {
		style = request.Voice.Style
	}

	// Adjust stability for v3 model - only accepts 0.0, 0.5, or 1.0
	stability := request.Voice.Stability
	if model == "eleven_v3" {
		if stability < 0.25 {
			stability = 0.0
		} else if stability < 0.75 {
			stability = 0.5
		} else {
			stability = 1.0
		}
	}

	logger("Using style: "+fmt.Sprintf("%f", style), logDebug, request.Channel)
	logger("Using stability: "+fmt.Sprintf("%f", stability), logDebug, request.Channel)

	// Channel to capture TTS errors from the goroutine
	errChan := make(chan error, 1)

	go func() {
		var err error
		err = ttsStream(ctx, elevenKey, pipeWriter, request.Text, model, request.Voice.Voice, stability, request.Voice.SimilarityBoost, style, request.Voice.Speed, request.Voice.UseSpeakerBoost, request.Voice.LanguageCode, format)
		if err != nil {
			// Log detailed parameters when API call fails
			voiceName, _ := getVoiceName(request.Voice.Voice)
			logger(fmt.Sprintf("Error generating TTS audio: %s | Parameters: text=%q, voice=%s (ID: %s), model=%s, stability=%.2f, similarity_boost=%.2f, format=%s",
				err.Error(), request.Text, voiceName, request.Voice.Voice, model, stability, request.Voice.SimilarityBoost, format), logError, request.Channel)
			errChan <- err
		} else {
			errChan <- nil
		}
		pipeWriter.Close()
	}()

	audioData, err := io.ReadAll(pipeReader)
	if err != nil {
		logger("Error reading TTS audio data: "+err.Error(), logError, request.Channel)
		return nil, err
	}

	// Check if there was a TTS generation error
	ttsErr := <-errChan
	if ttsErr != nil {
		return nil, ttsErr
	}

	// Check if audio data is empty (can happen with some API errors)
	if len(audioData) == 0 {
		voiceName, _ := getVoiceName(request.Voice.Voice)
		logger(fmt.Sprintf("Empty audio data received | Parameters: text=%q, voice=%s (ID: %s), stability=%.2f, similarity_boost=%.2f",
			request.Text, voiceName, request.Voice.Voice, request.Voice.Stability, request.Voice.SimilarityBoost), logError, request.Channel)
		return nil, fmt.Errorf("empty audio data received from TTS API")
	}

	if verb {
		verbAudio := reverb(audioData, request.Channel)
		if verbAudio == nil {
			logger("Error applying reverb to audio", logError, request.Channel)
			return nil, fmt.Errorf("error applying reverb to audio")
		}
		audioData = verbAudio
	}

	return audioData, nil
}

// ttsStream is a custom TTS function that handles all models.
// For v2 (eleven_multilingual_v2): includes style, speed, use_speaker_boost, and language_code.
// For v3/turbo/flash: excludes style, speed, use_speaker_boost, and language_code.
func ttsStream(ctx context.Context, apiKey string, w io.Writer, text, modelID, voiceID string, stability, clarity, style, speed float64, useSpeakerBoost *bool, languageCode, format string) error {
	url := "https://api.elevenlabs.io/v1/text-to-speech/" + voiceID + "/stream"

	voiceSettings := map[string]interface{}{
		"stability":        stability,
		"similarity_boost": clarity,
	}

	requestBody := map[string]interface{}{
		"text":           text,
		"model_id":       modelID,
		"output_format":  format,
		"voice_settings": voiceSettings,
	}

	// v2-only parameters
	if modelID == "eleven_multilingual_v2" {
		voiceSettings["style"] = style
		if speed != 0 {
			voiceSettings["speed"] = speed
		}
		if useSpeakerBoost != nil {
			voiceSettings["use_speaker_boost"] = *useSpeakerBoost
		}
		if languageCode != "" {
			requestBody["language_code"] = languageCode
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "audio/mpeg")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// getUserInfo fetches user info from the ElevenLabs API
func getUserInfo(ctx context.Context) (*elevenLabsUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.elevenlabs.io/v1/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("xi-api-key", elevenKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var userInfo elevenLabsUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// getVoicePreviewURL fetches the preview URL for a voice from the ElevenLabs API
func getVoicePreviewURL(ctx context.Context, voiceID string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.elevenlabs.io/v1/voices/"+voiceID, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("xi-api-key", elevenKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var voiceInfo elevenLabsVoiceInfo
	if err := json.NewDecoder(resp.Body).Decode(&voiceInfo); err != nil {
		return "", err
	}

	return voiceInfo.PreviewURL, nil
}

type ClientData struct {
	CharactersLeft  int `json:"characters_left"`
	CharactersReset int `json:"characters_reset"`
}

func getCharactersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	userInfo, err := getUserInfo(ctx)
	if err != nil {
		logger("Error getting user info: "+err.Error(), logError, "Universal")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	characters := userInfo.Subscription.CharacterCount
	characterLimit := userInfo.Subscription.CharacterLimit

	charactersRemaining := characterLimit - characters

	charactersReset := userInfo.Subscription.NextCharacterCountResetUnix

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
