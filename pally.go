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
	pallyKeys []PallyKeys
)

type PallyKeys struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

func setupPally() {
	keys := os.Getenv("PALLY_KEYS")
	err := json.Unmarshal([]byte(keys), &pallyKeys)
	if err != nil {
		logger("Error unmarshalling Pally keys: "+err.Error(), logError)
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

func connectToPallyWebsocket(channel string, pallyKey string) {
	for {
		if err := attemptConnectToPallyWebsocket(channel, pallyKey); err != nil {
			logger("We need to restart the Pally connection", logInfo)
			logger("Reconnecting to Pally...", logInfo)
		} else {
			logger("Pally connection closed normally.", logInfo)
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
	logger("Received message from Pally", logDebug)

	time_of_message := time.Now()

	var found bool

	var campaignTipNotify CampaignTipNotify
	err := json.Unmarshal(message, &campaignTipNotify)
	if err != nil {
		logger("Error unmarshalling message: "+err.Error(), logError)
	}
	notifyType := campaignTipNotify.Type
	if notifyType != "campaigntip.notify" {
		logger("Not a campaign tip notification", logDebug)
		return
	}
	username := campaignTipNotify.Payload.CampaignTip.DisplayName
	if username == "" {
		logger(fmt.Sprintf("%v", campaignTipNotify), logDebug)
		logger("No username found in message so assuming it's not a tip notification", logDebug)
		return
	}

	// Check if there is a connected client for the channel
	for {
		if time.Now().Sub(time_of_message) > 30*time.Second {
			logger("No connected client for channel", logInfo)
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
	if cents == 0 {
		amountFormatted = fmt.Sprintf("%d dollars", dollars)
	} else {
		amountFormatted = fmt.Sprintf("%d dollars and %02d cents", dollars, cents)
	}
	ttsMessage := campaignTipNotify.Payload.CampaignTip.Message
	if ttsMessage == "" {
		ttsMessage = fmt.Sprintf("%s just tipped %s to the mods!", username, amountFormatted)
	} else {
		ttsMessage = fmt.Sprintf("%s just tipped %s to the mods! %s", username, amountFormatted, ttsMessage)
	}
	requestTime := fmt.Sprintf("%d", time.Now().UnixNano())
	logger(ttsMessage, logInfo)
	request := Request{
		Channel: channel,
		Text:    ttsMessage,
		Time:    requestTime,
		Voice: TTSSettings{
			Voice:           defaultVoice,
			Stability:       0.40,
			SimilarityBoost: 1.00,
			Style:           0.00,
		},
	}

	// Add the request to the queue
	requests = append(requests, request)
	go handleTTSAudio(nil, nil, request, true)
}

func attemptConnectToPallyWebsocket(channel string, pallyKey string) error {
	logger("Connecting to Pally WebSocket on channel "+channel, logInfo)

	// Create the WebSocket URL
	url := fmt.Sprintf("wss://events.pally.gg?auth=%s&channel=firehose", pallyKey)

	// Create a new WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logger("Error connecting to Pally WebSocket on channel "+channel+": "+err.Error(), logError)
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
			logger("Sending ping message to Pally on channel "+channel, logFountain)
			err = conn.WriteMessage(websocket.TextMessage, []byte(`ping`))
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					logger("Stopping ping on old connection for Pally on channel "+channel, logInfo)
					return
				} else {
					logger("Error sending ping message to Pally on channel "+channel+": "+err.Error(), logError)
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
			if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
				logger("Pally connection closed normally on channel "+channel, logInfo)
				return err
			} else {
				logger("Error reading message from Pally on channel "+channel+": "+err.Error(), logError)
				return err
			}
		}

		// check for pong messages
		if string(message) == "pong" {
			logger("Received pong message from Pally on channel "+channel, logFountain)
			continue
		}

		go handlePallyMessage(message, channel)
	}
}
