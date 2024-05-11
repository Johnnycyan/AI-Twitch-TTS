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

	// Send periodic ping messages to the client
	go func(clientName string, channel string, conn *websocket.Conn) {
		pingTicker := time.NewTicker(30 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case <-pingTicker.C:
				if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "use of closed network connection") {
						logger("Client "+clientName+" disconnected from channel "+channel, logInfo)
					} else {
						logger("Error sending ping message: "+err.Error(), logError)
					}
					conn.Close()
					delete(clients, conn)
					//remove clientname from map
					mapMutex.Lock()
					delete(addrToNameMap, fmt.Sprintf("%p", conn))
					mapMutex.Unlock()
					return
				}
			}
		}
	}(clientName, channel, conn)

	// Read messages from the client
	go func(clientName string, channel string, conn *websocket.Conn) {
		clientPingTicker := time.NewTicker(60 * time.Second)
		// check for client ping messages and reset the ticker, otherwise close the connection if no ping is received after 60 seconds
		go func() {
			for {
				select {
				case <-clientPingTicker.C:
					logger("Ping not received, closing connection for client "+clientName+" on channel "+channel, logInfo)
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
					conn.Close()
					delete(clients, conn)
					//remove clientname from map
					mapMutex.Lock()
					delete(addrToNameMap, fmt.Sprintf("%p", conn))
					mapMutex.Unlock()
					return
				} else if message == "confirm" {
					logger("Client "+clientName+" confirmed playing audio on channel "+channel, logInfo)
				} else {
					logger("Unknown message from "+clientName+" on channel "+channel+": "+message, logDebug)
				}
			} else if messageType == websocket.BinaryMessage {
				logger("Received binary message from "+clientName+" on channel "+channel, logDebug)
			}
		}
	}(clientName, channel, conn)
}
