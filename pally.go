package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pallyKeys   []PallyKeys
	pallyVoices []PallyVoice
)

type PallyKeys struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type PallyVoice struct {
	Channel string `json:"channel"`
	Voice   string `json:"voice"`
}

func getPallyVoiceID(channel string) string {
	for _, voice := range pallyVoices {
		if voice.Channel == channel {
			pallyVoiceID, err := getVoiceID(voice.Voice)
			if err != nil {
				logger("Error getting voice ID: "+err.Error(), logError, channel)
				return defaultVoiceID
			}
			return pallyVoiceID
		}
	}
	return defaultVoiceID
}

func setupPally() {
	keys := os.Getenv("PALLY_KEYS")
	err := json.Unmarshal([]byte(keys), &pallyKeys)
	if err != nil {
		logger("Error unmarshalling Pally keys: "+err.Error(), logError, "Universal")
		return
	}
	for _, key := range pallyKeys {
		if key.Name == "" || key.Key == "" {
			continue
		} else {
			go connectToPallyWebsocket(key.Name, key.Key)
		}
	}
}

func setupPallyVoices() {
	voices := os.Getenv("PALLY_VOICES")
	err := json.Unmarshal([]byte(voices), &pallyVoices)
	if err != nil {
		logger("Error unmarshalling Pally voices: "+err.Error(), logError, "Universal")
		return
	}
}

func connectToPallyWebsocket(channel string, pallyKey string) {
	logger("Connecting to Pally WebSocket", logInfo, channel)
	for {
		if err := attemptConnectToPallyWebsocket(channel, pallyKey); err != nil {
			continue
		} else {
			logger("Pally connection closed normally.", logInfo, channel)
			return
		}
	}
}

type CampaignTipNotify struct {
	Type    string          `json:"type"`
	Payload CampaignTipData `json:"payload"`
}

type CampaignTipData struct {
	CampaignTip CampaignTip `json:"campaignTip"`
	Page        Page        `json:"page"`
}

type CampaignTip struct {
	CreatedAt            string `json:"createdAt"`
	DisplayName          string `json:"displayName"`
	GrossAmountInCents   int    `json:"grossAmountInCents"`
	ID                   string `json:"id"`
	Message              string `json:"message"`
	NetAmountInCents     int    `json:"netAmountInCents"`
	ProcessingFeeInCents int    `json:"processingFeeInCents"`
	UpdatedAt            string `json:"updatedAt"`
}

type Page struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

func handlePallyMessage(message []byte, channel string) {
	logger("Received message from Pally", logDebug, channel)

	var checkRequest []Request

	for _, request := range requests {
		if request.Channel == channel {
			checkRequest = append(checkRequest, request)
		}
	}

	if len(checkRequest) > 0 {
		logger("Last audio is still playing on "+channel, logInfo, channel)
		time.Sleep(1 * time.Second)
		go handlePallyMessage(message, channel)
		return
	}

	time_of_message := time.Now()

	var found bool

	var campaignTipNotify CampaignTipNotify
	err := json.Unmarshal(message, &campaignTipNotify)
	if err != nil {
		logger("Error unmarshalling message: "+err.Error(), logError, channel)
	}
	notifyType := campaignTipNotify.Type
	if notifyType != "campaigntip.notify" {
		logger("Not a campaign tip notification", logDebug, channel)
		return
	}
	username := campaignTipNotify.Payload.CampaignTip.DisplayName
	if username == "" {
		logger("No username found in message so assuming it's an anonymous tip", logInfo, channel)
		username = "Anonymous"
	}

	// Check if there is a connected client for the channel
	for {
		if time.Since(time_of_message) > 30*time.Second {
			logger("No connected client", logInfo, channel)
			found = false
			return
		}
		for _, clientChannel := range clients {
			if clientChannel == channel {
				found = true
				break
			}
		}
		if found {
			break
		}
		time.Sleep(1 * time.Second)
	}

	amount := campaignTipNotify.Payload.CampaignTip.GrossAmountInCents
	// format as "x dollars and xx cents" as a string. No $ sign. It will be added in the TTS message
	//if there are no cents, it will be formatted as "x dollars"
	dollars := amount / 100
	cents := amount % 100
	var amountFormatted string
	if cents == 0 && dollars == 1 {
		amountFormatted = fmt.Sprintf("%d dollar", dollars)
	} else if cents == 0 && dollars > 1 {
		amountFormatted = fmt.Sprintf("%d dollars", dollars)
	} else if cents == 1 && dollars == 0 {
		amountFormatted = fmt.Sprintf("%d cent", cents)
	} else if cents == 1 && dollars == 1 {
		amountFormatted = fmt.Sprintf("%d dollar and %d cent", dollars, cents)
	} else if cents == 1 && dollars > 1 {
		amountFormatted = fmt.Sprintf("%d dollars and %d cent", dollars, cents)
	} else if cents > 1 && dollars == 0 {
		amountFormatted = fmt.Sprintf("%d cents", cents)
	} else {
		amountFormatted = fmt.Sprintf("%d dollars and %02d cents", dollars, cents)
	}
	ttsMessage := campaignTipNotify.Payload.CampaignTip.Message
	if ttsMessage == "" {
		ttsMessage = fmt.Sprintf("%s just tipped %s to the mods!", username, amountFormatted)
	} else {
		ttsMessage = fmt.Sprintf("%s just tipped %s to the mods! %s", username, amountFormatted, ttsMessage)
	}
	// ttsMessage, err = convertNumberToWords(ttsMessage)
	// if err != nil {
	// 	logger("Error converting number to words: "+err.Error(), logError, channel)
	// }
	requestTime := fmt.Sprintf("%d", time.Now().UnixNano())
	logger(ttsMessage, logInfo, channel)
	voice := getPallyVoiceID(channel)
	style, err := getVoiceStyle(voice)
	if err != nil {
		style = 0.00
	}
	request := Request{
		Channel: channel,
		Text:    ttsMessage,
		Time:    requestTime,
		Voice: TTSSettings{
			Voice:           voice,
			Stability:       0.40,
			SimilarityBoost: 1.00,
			Style:           style,
		},
	}

	// Add the request to the queue
	requests = append(requests, request)
	go handleTTSAudio(nil, nil, request, true)

	if mongoEnabled {
		request.Channel = request.Channel + "-pally"
		data, err := createData(request)
		if err != nil {
			logger("Error creating data: "+err.Error(), logError, channel)
			return
		}
		addData(data)
	}
}

func attemptConnectToPallyWebsocket(channel string, pallyKey string) error {
	// Create the WebSocket URL
	url := fmt.Sprintf("wss://events.pally.gg?auth=%s&channel=firehose", pallyKey)

	// Create a new WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logger("Error connecting to Pally WebSocket: "+err.Error(), logError, channel)
		return err
	}
	defer conn.Close()

	// send test echo message
	// go func() {
	// 	for {
	// 		time.Sleep(10 * time.Second)
	// 		err = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"echo","payload":{"type":"campaigntip.notify","payload":{"campaignTip":{"createdAt":"2024-03-13T18:02:33.743Z","displayName":"Cisco","grossAmountInCents":500,"id":"b1w2pjwjtb9fx0v1se9ex4n2","message":"Hello","netAmountInCents":500,"processingFeeInCents":0,"updatedAt":"2024-03-13T18:02:33.743Z"},"page":{"id":"1627451579049x550722173620715500","slug":"pally","title":"Pally.gg's Team Page","url":"https://pally.gg/p/pally"}}}}`))
	// 		if err != nil {
	// 			log.Println("Error sending test message to Pally:", err)
	// 		}
	// 		time.Sleep(590 * time.Second)
	// 	}
	// }()

	// send ping message every 60 seconds
	go func() {
		for {
			time.Sleep(60 * time.Second)
			logger("Sending ping message to Pally", logFountain, channel)
			err = conn.WriteMessage(websocket.TextMessage, []byte(`ping`))
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "close sent") {
					return
				} else {
					logger("Error sending ping message to Pally: "+err.Error(), logError, channel)
					return
				}
			}
		}
	}()

	// Reconnect on ping failure or connection failure
	for {
		// Read message from WebSocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			// check if it's just an EOF 1006 error
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) || websocket.IsCloseError(err, websocket.CloseGoingAway) {
				logger("Pally connection closed, reconnecting", logInfo, channel)
				return err
			} else {
				logger("Error reading message from Pally: "+err.Error(), logError, channel)
				return err
			}
		}

		// check for pong messages
		if string(message) == "pong" {
			logger("Received pong message from Pally", logFountain, channel)
			continue
		}

		go handlePallyMessage(message, channel)
	}
}
