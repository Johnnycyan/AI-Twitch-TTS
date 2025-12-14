package main

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
)

// CreatePageData holds all data for the create page template
type CreatePageData struct {
	Voices    []VoicePreview
	Effects   []string
	Modifiers []string
	Tags      []string
}

// handleCreate serves the TTS message creator page
func handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Surrogate-Control", "no-store")

	// Get voice previews
	var voicePreviews []VoicePreview
	for voice := range voices {
		voiceInfo, err := ttsClient.GetVoice(ctx, voices[voice].ID)
		if err != nil {
			logger("Error getting voice preview: "+err.Error(), logError, "Universal")
			continue
		}
		voicePreviews = append(voicePreviews, VoicePreview{
			Name:       voices[voice].Name,
			PreviewURL: voiceInfo.PreviewURL,
		})
	}

	// Sort voices alphabetically
	sort.Slice(voicePreviews, func(i, j int) bool {
		return voicePreviews[i].Name < voicePreviews[j].Name
	})

	// Get effects list
	effects := getEffectsList()

	// Prepare template data
	data := CreatePageData{
		Voices:    voicePreviews,
		Effects:   effects,
		Modifiers: []string{"reverb"},
		Tags: []string{
			"laughter",
			"laughs",
			"sad",
			"sigh",
			"cries",
			"screams",
			"gasps",
			"groans",
			"sniffs",
		},
	}

	tmpl, err := template.ParseFiles("static/create.html")
	if err != nil {
		logger("Error parsing create template: "+err.Error(), logError, "Universal")
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger("Error executing create template: "+err.Error(), logError, "Universal")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getEffectsList returns a sorted list of effect names
func getEffectsList() []string {
	var effects []string

	if _, err := os.Stat(effectFolder); err != nil {
		return effects
	}

	files, err := os.ReadDir(effectFolder)
	if err != nil {
		logger("Error reading effects folder: "+err.Error(), logError, "Universal")
		return effects
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mp3") {
			effects = append(effects, strings.TrimSuffix(file.Name(), ".mp3"))
		}
	}

	slices.Sort(effects)
	return effects
}
