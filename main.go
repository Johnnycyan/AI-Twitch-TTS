package main

import (
	"html/template"
	"net/http"
)

var (
	port = "8034"
)

func setupHandlers() {
	http.HandleFunc("/tts", handleTTS)
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", serveClient)
}

func main() {
	setupENV()
	setupHandlers()
	logger("Server listening on port: "+port, logInfo)
	http.ListenAndServe(":"+port, nil)
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
