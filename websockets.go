package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gorilla/websocket"
)

var (
	clients       = make(map[*websocket.Conn]string)
	addrToNameMap = make(map[string]string)
	mapMutex      = sync.Mutex{}
	playing       = make(map[string]bool)
)

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

func clearChannelRequests(channel string) {
	defer func() {
		if r := recover(); r != nil {
			requests = nil
			logger("Recovered from panic in clearChannelRequests: "+fmt.Sprintf("%v", r), logError)
		}
	}()
	for i := len(requests) - 1; i >= 0; i-- {
		if i >= len(requests) {
			continue
		}
		request := requests[i]
		if request.Channel == channel {
			requests = append(requests[:i], requests[i+1:]...)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger("Error upgrading to WebSocket: "+err.Error(), logError)
		return
	}
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	clientName := getClientName(fmt.Sprintf("%p", conn))
	logger("Client "+clientName+" connected to channel "+channel, logInfo)
	clients[conn] = channel

	// Read messages from the client
	go func(clientName string, channel string, conn *websocket.Conn) {
		clientPingTicker := time.NewTicker(60 * time.Second)
		// check for client ping messages and reset the ticker, otherwise close the connection if no ping is received after 60 seconds
		go func() {
			for {
				select {
				case <-clientPingTicker.C:
					logger("Ping not received, closing connection for client "+clientName+" on channel "+channel, logInfo)
					clearChannelRequests(channel)
					conn.Close()
					delete(clients, conn)
					//remove clientname from map
					mapMutex.Lock()
					delete(addrToNameMap, fmt.Sprintf("%p", conn))
					mapMutex.Unlock()
					return
				}
			}
		}()
		defer clientPingTicker.Stop()

		for {
			messageType, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					logger("Client "+clientName+" disconnected from channel "+channel, logInfo)
				} else {
					logger("Error reading message from client "+clientName+" on channel "+channel+": "+err.Error(), logError)
				}
				conn.Close()
				delete(clients, conn)
				//remove clientname from map
				mapMutex.Lock()
				delete(addrToNameMap, fmt.Sprintf("%p", conn))
				mapMutex.Unlock()
				return
			}

			if messageType == websocket.TextMessage {
				message := string(messageBytes)
				if message == "ping" {
					logger("Received ping from "+clientName+" on channel "+channel, logFountain)
					clientPingTicker.Reset(60 * time.Second)
				} else if message == "close" {
					logger("Client "+clientName+" closed the connection on channel "+channel, logInfo)
					clearChannelRequests(channel)
					conn.Close()
					delete(clients, conn)
					//remove clientname from map
					mapMutex.Lock()
					delete(addrToNameMap, fmt.Sprintf("%p", conn))
					mapMutex.Unlock()
					return
				} else if strings.Contains(message, "confirm") {
					// split the message by spaces and get the last element
					// this is the timestamp of the audio that the client is confirming
					timestamp := strings.Split(message, " ")[1]
					logger("Client "+clientName+" confirmed playing audio for "+timestamp+" on channel "+channel, logInfo)
					// remove timestamp from playing map
					delete(playing, timestamp)
				} else {
					logger("Unknown message from "+clientName+" on channel "+channel+": "+message, logDebug)
				}
			} else if messageType == websocket.BinaryMessage {
				logger("Received binary message from "+clientName+" on channel "+channel, logDebug)
			}
		}
	}(clientName, channel, conn)
}

func sendTextMessage(channel string, message string) {
	for client, clientChannel := range clients {
		if clientChannel == channel {
			clientName := getClientName(fmt.Sprintf("%p", client))
			err := client.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				logger("Error sending text message to "+clientName+": "+err.Error(), logError)
				client.Close()
				delete(clients, client)
			}
			logger("Text message sent to "+clientName+" on channel "+channel, logInfo)
		}
	}
}
