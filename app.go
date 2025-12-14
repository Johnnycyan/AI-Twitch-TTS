package main

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
)

// VoiceData represents voice data for the API
type VoiceData struct {
	Name       string `json:"name"`
	PreviewURL string `json:"preview_url"`
}

// handleApp serves the SPA application
func handleApp(w http.ResponseWriter, r *http.Request) {
	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	tmpl, err := template.ParseFiles("static/app.html")
	if err != nil {
		logger("Error parsing app template: "+err.Error(), logError, "Universal")
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		logger("Error executing app template: "+err.Error(), logError, "Universal")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleAPIVoices returns voices as JSON for the SPA
func handleAPIVoices(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	w.Header().Set("Content-Type", "application/json")

	var voiceList []VoiceData
	for _, v := range voices {
		voiceInfo, err := ttsClient.GetVoice(ctx, v.ID)
		if err != nil {
			continue
		}
		voiceList = append(voiceList, VoiceData{
			Name:       v.Name,
			PreviewURL: voiceInfo.PreviewURL,
		})
	}

	// Sort alphabetically
	sort.Slice(voiceList, func(i, j int) bool {
		return voiceList[i].Name < voiceList[j].Name
	})

	json.NewEncoder(w).Encode(voiceList)
}

// handleAPIEffects returns effects as JSON for the SPA
func handleAPIEffects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var effects []string

	if _, err := os.Stat(effectFolder); err != nil {
		json.NewEncoder(w).Encode(effects)
		return
	}

	files, err := os.ReadDir(effectFolder)
	if err != nil {
		json.NewEncoder(w).Encode(effects)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mp3") {
			effects = append(effects, strings.TrimSuffix(file.Name(), ".mp3"))
		}
	}

	slices.Sort(effects)
	json.NewEncoder(w).Encode(effects)
}
