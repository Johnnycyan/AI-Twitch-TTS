package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sandisuryadi36/number-to-words/convert"
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

func convertNumberToWords(text string) (string, error) {
	logger("Converting numbers to words", logDebug, "Universal")
	// find numbers in the string, convert them to words and replace them in the string. I want to find them even if it's for example xdd34624 so not only numbers that are separated by spaces
	// I'm using a regex to find the numbers and then convert them to words

	// find all numbers in the string
	re := regexp.MustCompile(`\d+(\.\d+)?`) // example: 123, 123.48, xdd33444 4234xdd -> 123, 123.48, 33444 4234
	numbers := re.FindAllString(text, -1)

	for _, num := range numbers {
		for strings.HasPrefix(num, "0") {
			// remove the 0 from the number
			num = num[1:]
			logger("Removing 0 from the number", logDebug, "Universal")
		}
	}

	if numbers == nil {
		logger("No numbers found in the text", logDebug, "Universal")
		return text, nil
	}
	logger("Numbers found: "+fmt.Sprintf("%+v", numbers), logDebug, "Universal")

	// add a space between any string character and number
	text = re.ReplaceAllString(text, " $0") // example: 123, 123.48, xdd33444 -> 123, 123.48, xdd 33444

	// convert the numbers to words
	var loopErr error
	originalText := text
	for _, number := range numbers {
		defer func() {
			if r := recover(); r != nil {
				loopErr = fmt.Errorf("Error converting number to words: %v", r)
				logger(loopErr.Error(), logError, "Universal")
				return
			}
		}()
		// convert the number to words
		words := convert.NumberToWords(number, "en")

		// replace the number in the string with the words
		text = strings.Replace(text, number, words, -1)
	}

	if loopErr != nil {
		return originalText, loopErr
	}

	text = strings.TrimSpace(text)
	text = strings.Replace(text, "   ", " ", -1)
	text = strings.Replace(text, "  ", " ", -1)

	return text, nil
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

func addPartsToRequest(parts []Part, requestTime string, params *URLParams) error {
	for _, part := range parts {
		if part.Type == "text" {
			voice, err := getVoiceID(params.FallbackVoice)
			if err != nil {
				logger("Error getting voice ID: "+err.Error(), logError, params.Channel)
				if len(requests) > 0 {
					clearChannelRequests(params.Channel)
				}
				return fmt.Errorf("Invalid fallback voice")
			}
			// fixedText, err := convertNumberToWords(part.Text)
			// if err != nil {
			// 	logger("Error converting number to words: "+err.Error(), logError, params.Channel)
			// 	fixedText = part.Text
			// }
			fixedText := part.Text
			requests = append(requests, Request{
				Index:   len(requests) + 1,
				Type:    part.Type,
				Channel: params.Channel,
				Time:    requestTime,
				Params:  *params,
				Voice: TTSSettings{
					Voice:           voice,
					Stability:       params.Stability,
					SimilarityBoost: params.SimilarityBoost,
					Style:           params.Style,
				},
				Text: fixedText,
			})
		} else if part.Type == "voice" {
			voice, err := getVoiceID(part.Voice)
			if err != nil {
				logger("Error getting voice ID: "+err.Error(), logError, params.Channel)
				if len(requests) > 0 {
					clearChannelRequests(params.Channel)
				}
				return fmt.Errorf("You used an invalid voice tag: " + part.Voice)
			}
			// fixedText, err := convertNumberToWords(part.Text)
			// if err != nil {
			// 	logger("Error converting number to words: "+err.Error(), logError, params.Channel)
			// 	fixedText = part.Text
			// }
			fixedText := part.Text
			requests = append(requests, Request{
				Index:   len(requests) + 1,
				Type:    part.Type,
				Channel: params.Channel,
				Time:    requestTime,
				Params:  *params,
				Voice: TTSSettings{
					Voice:           voice,
					Stability:       params.Stability,
					SimilarityBoost: params.SimilarityBoost,
					Style:           params.Style,
				},
				Text: fixedText,
			})
		} else if part.Type == "effect" {
			requests = append(requests, Request{
				Index:   len(requests) + 1,
				Type:    part.Type,
				Channel: params.Channel,
				Time:    requestTime,
				Params:  *params,
				Effect:  part.Effect,
			})
		}
	}
	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	requestTime := fmt.Sprintf("%d", time.Now().UnixNano())

	params := getURLParams(r)

	defer func(channel string) {
		if r := recover(); r != nil {
			logger("Recovered from panic in handleRequest: "+fmt.Sprintf("%v", r), logError, channel)
			http.Error(w, "Error processing request. Maybe you used an unsupported symbol?", http.StatusInternalServerError)
		}
	}(params.Channel)

	logger("Received Audio request", logInfo, params.Channel)

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
		if len(checkRequest) > 0 {
			clearChannelRequests(params.Channel)
		}
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
		if len(requests) > 0 {
			clearChannelRequests(params.Channel)
		}
		return
	}

	if !validVoice(params.FallbackVoice) {
		logger("Invalid fallback voice: "+params.FallbackVoice, logInfo, params.Channel)
		params.FallbackVoice = defaultVoice
	}

	logger("Getting text parts", logDebug, params.Channel)
	messages, err := getTextParts(params.Text)
	if err != nil {
		logger("Error getting text parts: "+err.Error(), logError, params.Channel)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger("Getting formatted parts", logDebug, params.Channel)
	parts, err := getFormattedParts(messages)
	if err != nil {
		logger("Error getting formatted parts: "+err.Error(), logError, params.Channel)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger("Adding parts to request", logDebug, params.Channel)
	err = addPartsToRequest(parts, requestTime, params)
	if err != nil {
		logger("Error adding parts to request: "+err.Error(), logError, params.Channel)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	processRequest(w, r, params)
}

func getTextParts(text string) ([]string, error) {
	// check if message starts with a tag
	if !strings.HasPrefix(text, "[") {
		if strings.Contains(text, "[e-") || strings.Contains(text, "[v-") {
			// if it doesn't start with a tag but contains a tag, this is an error
			logger("Text contains a tag but doesn't start with a tag: "+text, logError, "Universal")
			return nil, fmt.Errorf("If you use any tags, the text must start with a tag.")
		}
		return []string{text}, nil
	}
	// Compile a regular expression to find the tags and the following text
	re := regexp.MustCompile(`(\[[^\]]+\][^[]*)`)

	// Find all matches
	matches := re.FindAllString(text, -1)

	// If no matches are found, return the original text as a single-element slice
	if len(matches) == 0 {
		return []string{text}, nil
	}

	return matches, nil
}

func getFormattedParts(parts []string) ([]Part, error) {
	var result []Part

	for _, part := range parts {
		// Check for a voice or effect tag at the beginning of the part
		tagRe := regexp.MustCompile(`^\[([^\]]+)\]`)
		tagMatch := tagRe.FindStringSubmatch(part)

		if tagMatch != nil {
			tag := tagMatch[1]
			text := strings.TrimSpace(part[len(tagMatch[0]):])

			if strings.HasPrefix(tag, "v-") {
				// It's a voice tag, remove the "v-" prefix
				result = append(result, Part{
					Type:  "voice",
					Text:  text,
					Voice: tag[2:], // Remove "v-" prefix
				})
			} else if strings.HasPrefix(tag, "e-") {
				if text != "" {
					// It's an effect tag with text, this shouldn't happen so return an error
					return nil, fmt.Errorf("Looks like you might have missed a voice tag. Affected text: " + text)
				}
				// It's an effect tag, remove the "e-" prefix
				result = append(result, Part{
					Type:   "effect",
					Effect: tag[2:], // Remove "e-" prefix
				})
			} else {
				// Unknown tag, this is an error
				return nil, fmt.Errorf("Unknown tag. Not a voice tag or effect tag: " + tag)
			}
		} else {
			// No tag found, treat as plain text (fallback scenario)
			result = append(result, Part{
				Type: "text",
				Text: part,
			})
		}
	}

	return result, nil
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

type AudioData struct {
	Audio []byte
}

func processRequest(w http.ResponseWriter, _ *http.Request, params *URLParams) {
	logger("Processing request", logInfo, params.Channel)
	var audio []byte
	var audioData []AudioData
	var err error
	var bad = false
	for _, request := range requests {
		if request.Type == "text" || request.Type == "voice" {
			audio, err = generateAudio(request)
			if err != nil {
				logger("Error generating audio: "+err.Error(), logError, params.Channel)
				bad = true
				if len(requests) > 0 {
					clearChannelRequests(params.Channel)
				}
				http.Error(w, "Error generating audio. Check your inputs.", http.StatusInternalServerError)
				return
			}
			if audio == nil || len(audio) == 0 {
				logger("No audio data generated", logError, params.Channel)
				bad = true
				if len(requests) > 0 {
					clearChannelRequests(params.Channel)
				}
				http.Error(w, "Error getting audio data. Check your inputs.", http.StatusInternalServerError)
				return
			}
			audioData = append(audioData, AudioData{Audio: audio})
			if mongoEnabled {
				data, err := createData(request)
				if err != nil {
					logger("Error creating data: "+err.Error(), logError, params.Channel)
					bad = true
					if len(requests) > 0 {
						clearChannelRequests(params.Channel)
					}
					http.Error(w, "Error creating data. Check your inputs.", http.StatusInternalServerError)
					return
				}
				addData(data)
			}
		} else if request.Type == "effect" {
			audio, found := getEffectSound(request.Effect)
			if !found {
				logger("Effect sound not found", logError, params.Channel)
				bad = true
				if len(requests) > 0 {
					clearChannelRequests(params.Channel)
				}
				http.Error(w, "You specified an effect that doesn't exist: "+request.Effect, http.StatusBadRequest)
				return
			}
			audioData = append(audioData, AudioData{Audio: audio})
			// go handleEffect(w, r, request)
		}
	}

	replyVerifyTicker := time.NewTicker(120 * time.Second)
	if !bad {
		for i, data := range audioData {
			playing[requests[i].Time] = true
			replyVerifyTicker.Reset(120 * time.Second)
			sendAudio(requests[i], data.Audio)
			for playing[requests[i].Time] {
				select {
				case <-replyVerifyTicker.C:
					requestName := getAudioDataName(requests[i].Time)
					logger("No reply received for "+requestName, logInfo, requests[i].Channel)
					clearChannelRequests(requests[i].Channel)
					http.Error(w, "No reply received for "+requestName, http.StatusRequestTimeout)
					sendTextMessage(requests[i].Channel, "reload")
					return
				default:
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
	} else {
		logger("Error processing request", logError, params.Channel)
		if len(requests) > 0 {
			clearChannelRequests(params.Channel)
		}
		http.Error(w, "Error processing request. Check your inputs.", http.StatusInternalServerError)
	}

	if len(requests) > 0 {
		clearChannelRequests(params.Channel)
		//remove audio data name from map
		audioMutex.Lock()
		for _, request := range requests {
			if request.Channel == params.Channel {
				delete(audioDataNames, request.Time)
			}
		}
		audioMutex.Unlock()
	}
}
