package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/getcohesive/elevenlabs/client"
	"github.com/getcohesive/elevenlabs/client/types"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	elevenKey     string
	clients       = make(map[*websocket.Conn]string)
	requestTime   = map[string]time.Time{}
	voices        = map[string]string{"morgan": "dCJWUtGAvzXajvDoIJdj", "chris": "G17SuINrv2H9FC6nvetn", "stanley": "ARh3OFuUWNL07IeeFwD5"}
	voice         = voices["chris"]
	pallyKey      string
	alertFolder   = "alerts"
	port          = "8034"
	logInfo       = "info"
	logDebug      = "debug"
	logError      = "error"
	logFountain   = "fountain"
	addrToNameMap = make(map[string]string)
	mapMutex      = sync.Mutex{}
	serverURL     string
	pallyChannel  string
	ttsKey        string
)

func logger(message string, level string) {
	args := os.Args
	var logLevel string
	if len(args) > 1 {
		logLevel = args[1]
	} else {
		logLevel = "debug"
	}

	switch level {
	case "error":
		message = "ERROR: " + message
	case "info":
		message = "INFO: " + message
	case "debug":
		message = "DEBUG: " + message
	case "fountain":
		message = "FOUNTAIN: " + message
	default:
		message = "UNKNOWN: " + message
	}

	switch logLevel {
	case "fountain":
		log.Println(message)
	case "debug":
		if level != "fountain" {
			log.Println(message)
		}
	case "info":
		if level != "debug" && level != "fountain" {
			log.Println(message)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logger("Error loading .env file", logError)
	}
	elevenKey = os.Getenv("ELEVENLABS_KEY")
	pallyKey = os.Getenv("PALLY_KEY")
	serverURL = os.Getenv("SERVER_URL")
	pallyChannel = os.Getenv("PALLY_CHANNEL")
	ttsKey = os.Getenv("TTS_KEY")
	go connectToPallyWebsocket()
	http.HandleFunc("/tts", handleTTS)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", serveClient)
	logger("Server listening on port: "+port, logInfo)
	http.ListenAndServe(":"+port, nil)
}

func getAlertSound() (*os.File, bool) {
	var alertSounds []string
	//check if alert folder exists and if so get all .mp3 files in it
	if _, err := os.Stat(alertFolder); err == nil {
		files, err := os.ReadDir(alertFolder)
		if err != nil {
			logger("Error reading alert folder: "+err.Error(), logError)
			return nil, false
		} else {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".mp3") {
					alertSounds = append(alertSounds, file.Name())
				}
			}
		}
	} else {
		return nil, false
	}

	// check if there are any alert sounds in the folder
	if len(alertSounds) == 0 {
		logger("No alert sounds found in alert folder: "+alertFolder, logDebug)
		return nil, false
	}

	// get a random alert sound from the list of alert sounds
	randomAlertSound := alertSounds[rand.Intn(len(alertSounds))]
	logger("Random alert sound selected: "+randomAlertSound, logDebug)
	alertSound, err := os.Open(fmt.Sprintf("%s/%s", alertFolder, randomAlertSound))
	if err != nil {
		logger("Error opening alert sound: "+err.Error(), logError)
		return nil, false
	}

	return alertSound, true
}

func handleTTSAudio(w http.ResponseWriter, _ *http.Request, text string, channel string, alert bool) {
	audioData, err := generateAudio(text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if alert {
		alertSound, alertExists := getAlertSound()

		if alertExists {
			alertSoundBytes, err := io.ReadAll(alertSound)
			if err != nil {
				logger("Error reading alert sound: "+err.Error(), logError)
			} else {
				for client, clientChannel := range clients {
					if clientChannel == channel {
						err := client.WriteMessage(websocket.BinaryMessage, alertSoundBytes)
						if err != nil {
							logger("Error sending alert sound to client: "+err.Error(), logError)
							client.Close()
							delete(clients, client)
						}
					}
				}
				logger("Alert sound sent to channel: "+channel, logInfo)
				time.Sleep(3 * time.Second)
			}
		}
	}

	for client, clientChannel := range clients {
		if clientChannel == channel {
			err := client.WriteMessage(websocket.BinaryMessage, audioData)
			if err != nil {
				logger("Error sending audio data to client: "+err.Error(), logError)
				client.Close()
				delete(clients, client)
			}
			logger("Audio data sent to channel: "+channel, logInfo)
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
		voice = voices["chris"]
	}
	// check if the voice is valid
	if _, ok := voices[voice]; !ok {
		logger("Invalid voice: "+voice+" so defaulting to Chris.", logInfo)
		voice = voices["chris"]
	} else {
		logger("Voice selected: "+voice, logDebug)
		voice = voices[voice]
	}

	if len(clients) == 0 {
		logger("No connected clients", logInfo)
		http.Error(w, "No connected clients", http.StatusNotFound)
		return
	}

	// Check if there is a connected client for the channel
	for _, clientChannel := range clients {
		if clientChannel == channel {
			logger("Connected client for channel: "+channel, logDebug)
			break
		}
		logger("No connected client for channel: "+channel, logInfo)
		http.Error(w, "No connected client for channel", http.StatusNotFound)
		return
	}

	if time.Since(requestTime[channel]) < 10*time.Second {
		logger("Rate limit exceeded for channel: "+channel, logInfo)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	requestTime[channel] = time.Now()

	go handleTTSAudio(w, r, text, channel, false)
	return
}

func generateAudio(text string) ([]byte, error) {
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
		err := client.TTSStream(ctx, pipeWriter, text, "eleven_multilingual_v2", voice, types.SynthesisOptions{Stability: 0.40, SimilarityBoost: 1.00, Format: format, Style: 0.25})
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

func generateRandomName() string {
	// Generate two random words
	adjective := randomdata.Adjective()
	noun := randomdata.Noun()
	return fmt.Sprintf("%s-%s", adjective, noun)
}

func getClientName(remoteAddr string) string {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	// Check if the name already exists
	name, exists := addrToNameMap[remoteAddr]
	if !exists {
		// Generate a new name if not exist
		name = generateRandomName()
		addrToNameMap[remoteAddr] = name
	}

	return name
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger("Error upgrading to WebSocket: "+err.Error(), logError)
		return
	}
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	clientName := getClientName(r.RemoteAddr)
	logger("Client "+clientName+" connected to channel: "+channel, logInfo)
	clients[conn] = channel

	// Send periodic ping messages to the client
	go func(clientName string, channel string, conn *websocket.Conn) {
		pingTicker := time.NewTicker(10 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case <-pingTicker.C:
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					if strings.Contains(err.Error(), "broken pipe") {
						logger("Client "+clientName+" disconnected from channel: "+channel, logInfo)
					} else {
						logger("Error sending ping message: "+err.Error(), logError)
					}
					conn.Close()
					delete(clients, conn)
					return
				}
			}
		}
	}(clientName, channel, conn)
}

func serveClient(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ServerURL string
	}{
		ServerURL: serverURL,
	}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		logger("Error parsing template: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		logger("Error executing template: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func connectToPallyWebsocket() {
	for {
		if err := attemptConnectToPallyWebsocket(); err != nil {
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

func handlePallyMessage(message []byte) {
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
			if clientChannel == pallyChannel {
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
	logger(ttsMessage, logInfo)
	go handleTTSAudio(nil, nil, ttsMessage, pallyChannel, true)
}

func attemptConnectToPallyWebsocket() error {
	logger("Connecting to Pally WebSocket...", logInfo)

	// Create the WebSocket URL
	url := fmt.Sprintf("wss://events.pally.gg?auth=%s&channel=firehose", pallyKey)

	// Create a new WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		logger("Error connecting to Pally WebSocket: "+err.Error(), logError)
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
			logger("Sending ping message to Pally", logFountain)
			err = conn.WriteMessage(websocket.TextMessage, []byte(`ping`))
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					logger("Stopping ping on old connection.", logInfo)
					return
				} else {
					logger("Error sending ping message to Pally: "+err.Error(), logError)
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
				logger("Pally connection closed normally.", logInfo)
				return err
			} else {
				logger("Error reading message from Pally: "+err.Error(), logError)
				return err
			}
		}

		// check for pong messages
		if string(message) == "pong" {
			logger("Received pong message from Pally", logFountain)
			continue
		}

		go handlePallyMessage(message)
	}
}
