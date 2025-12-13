package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"net/http"

	"github.com/gorilla/websocket"
)

var (
	requests       []Request
	audioDataNames = make(map[string]string)
	audioMutex     = sync.Mutex{}
)

type URLParams struct {
	Channel         string
	AuthKey         string
	Text            string
	FallbackVoice   string
	Stability       float64
	SimilarityBoost float64
	Style           float64
}

// Index: Index of the request, Type: Type of the request, Time: Time of the request, Params: URL parameters, Voice: TTS settings, Text: Text to be converted to speech
type Request struct {
	Index   int
	Type    string
	Channel string
	Time    string
	Params  URLParams
	Voice   TTSSettings
	Text    string
	Effect  string
}

type Part struct {
	Type   string
	Text   string
	Voice  string
	Effect string
}

func generateRandomAudioDataName() string {
	// Generate two random words
	color := strings.ToLower(randomColor())
	space := strings.ToLower(randomSpace())
	return fmt.Sprintf("%s-%s", color, space)
}

func getAudioDataName(audioData string) string {
	audioMutex.Lock()
	defer audioMutex.Unlock()

	// Check if the name already exists
	name, exists := audioDataNames[audioData]
	if !exists {
		// Generate a new name if not exist
		name = generateRandomAudioDataName()
		audioDataNames[audioData] = name
	}

	return name
}

func getURLParams(r *http.Request) *URLParams {
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	if channel == "" {
		return nil
	}

	logger("Getting URL parameters", logDebug, channel)

	authKey := r.URL.Query().Get("key")
	if authKey == "" {
		return nil
	}

	text := r.URL.Query().Get("text")
	if text == "" {
		return nil
	}

	fallbackVoice := strings.ToLower(r.URL.Query().Get("voice"))

	stabilityString := r.URL.Query().Get("stability")
	if stabilityString == "" {
		stabilityString = "0.40"
	}
	stability, err := strconv.ParseFloat(stabilityString, 64)
	if err != nil {
		return nil
	}

	similarityBoostString := r.URL.Query().Get("similarityBoost")
	if similarityBoostString == "" {
		similarityBoostString = "1.00"
	}
	similarityBoost, err := strconv.ParseFloat(similarityBoostString, 64)
	if err != nil {
		return nil
	}

	styleString := r.URL.Query().Get("style")
	if styleString == "" {
		styleString = "0.00"
	}
	style, err := strconv.ParseFloat(styleString, 64)
	if err != nil {
		return nil
	}

	params := &URLParams{
		Channel:         channel,
		AuthKey:         authKey,
		Text:            text,
		FallbackVoice:   fallbackVoice,
		Stability:       stability,
		SimilarityBoost: similarityBoost,
		Style:           style,
	}

	return params
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	params := getURLParams(r)

	defer func(channel string) {
		if r := recover(); r != nil {
			logger("Recovered from panic in handleRequest: "+fmt.Sprintf("%v", r), logError, channel)
			http.Error(w, "Error processing request. Maybe you used an unsupported symbol?", http.StatusInternalServerError)
		}
	}(params.Channel)

	logger("Received Audio request", logInfo, params.Channel)

	// Check if there's already a request being processed for this channel
	var checkRequest []Request
	for _, request := range requests {
		if request.Channel == params.Channel {
			checkRequest = append(checkRequest, request)
		}
	}

	if len(checkRequest) > 0 {
		logger("Last audio is still playing", logInfo, checkRequest[0].Channel)
		http.Error(w, "Wait for the last audio to finish playing", http.StatusTooManyRequests)
		return
	}

	if len(clients) == 0 {
		logger("No connected clients", logInfo, params.Channel)
		http.Error(w, "No connected clients", http.StatusNotFound)
		return
	}

	// Check if there is a connected client for the channel
	found := false
	for client, clientChannel := range clients {
		clientName := getClientName(fmt.Sprintf("%p", client))
		if clientChannel == params.Channel {
			logger("Found client "+clientName, logDebug, params.Channel)
			found = true
			break
		}
	}
	if !found {
		logger("No connected client", logInfo, params.Channel)
		http.Error(w, "No connected client for channel", http.StatusNotFound)
		return
	}

	// Use unified message processor
	msg := Message{
		Channel:         params.Channel,
		Text:            params.Text,
		DefaultVoice:    params.FallbackVoice,
		Stability:       params.Stability,
		SimilarityBoost: params.SimilarityBoost,
		Style:           params.Style,
		PlayAlert:       false,
	}

	err := ProcessAndPlay(msg)
	if err != nil {
		logger("Error processing request: "+err.Error(), logError, params.Channel)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func sendAudio(request Request, audioData []byte) {
	sendTextMessage(request.Channel, "start "+request.Time)
	time.Sleep(50 * time.Millisecond)
	requestName := getAudioDataName(request.Time)
	for client, clientChannel := range clients {
		if clientChannel == request.Channel {
			clientName := getClientName(fmt.Sprintf("%p", client))
			err := client.WriteMessage(websocket.BinaryMessage, audioData)
			if err != nil {
				logger("Error sending audio data to "+requestName+": "+err.Error(), logError, request.Channel)
				client.Close()
				delete(clients, client)
				if len(requests) > 0 {
					clearChannelRequests(request.Channel)
				}
			}
			logger("Audio data "+requestName+" sent to "+clientName, logInfo, request.Channel)
		}
	}
}