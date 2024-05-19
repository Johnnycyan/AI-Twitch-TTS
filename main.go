package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"os"

	//"log"
	"net/http"
	//"os"
)

var (
	port = "8039"
)

func setupHandlers() {
	// Serve the effects directory under the /effects route
	http.Handle("/effects/", http.StripPrefix("/effects", http.FileServer(http.Dir(effectFolder))))
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/voices", listVoices)
	http.HandleFunc("/effects", effectsHandler)
	http.HandleFunc("/tts", handleRequest)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/fx", listEffects)
	http.HandleFunc("/", serveClient)
	http.HandleFunc("/update", updateHandler)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	hash, err := ComputeMD5("index.html")
	if err != nil {
		logger("Error computing hash for index.html: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sendTextMessage(channel, "update "+hash)
}

func main() {
	// args := os.Args
	// if len(args) > 1 {
	// 	port = args[1]
	// } else {
	// 	log.Fatal("Port not provided")
	// }
	setupENV()
	setupHandlers()

	logger("Server listening on port: "+port, logInfo)
	http.ListenAndServe(":"+port, nil)
}

func ComputeMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func listVoices(w http.ResponseWriter, _ *http.Request) {
	var responseString string
	for i, voice := range voices {
		if i == len(voices)-1 {
			responseString += voice.Name
		} else {
			responseString += voice.Name + ", "
		}
	}

	w.Write([]byte(responseString))
}

func serveClient(w http.ResponseWriter, r *http.Request) {
	htmlHash, err := ComputeMD5("index.html")
	var data interface{}
	if sentryURL == "" {
		data = struct {
			ServerURL string
			Hash      string
		}{
			ServerURL: serverURL,
			Hash:      htmlHash,
		}
	} else {
		data = struct {
			ServerURL string
			SentryURL string
			Hash      string
		}{
			ServerURL: serverURL,
			SentryURL: sentryURL,
			Hash:      htmlHash,
		}
	}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		logger("Error parsing template: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Surrogate-Control", "no-store")

	err = tmpl.Execute(w, data)
	if err != nil {
		logger("Error executing template: "+err.Error(), logError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
