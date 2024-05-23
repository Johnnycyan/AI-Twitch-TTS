package main

import (
	"context"
	"html/template"
	"sort"

	"net/http"
)

type VoicePreview struct {
	Name       string `json:"name"`
	PreviewURL string `json:"preview_url"`
}

func listVoices(w http.ResponseWriter, _ *http.Request) {
	var responseString string
	var voiceList []string
	for voice := range voices {
		voiceList = append(voiceList, voices[voice].Name)
	}

	sort.Strings(voiceList)

	for i, voice := range voiceList {
		if i == len(voices)-1 {
			responseString += voice
		} else {
			responseString += voice + ", "
		}
	}

	w.Write([]byte(responseString))
}

func handleVoices(w http.ResponseWriter, r *http.Request) {
	list := r.URL.Query().Get("list")
	if list == "true" {
		listVoices(w, r)
		return
	}
	ctx := context.Background()
	var previews []VoicePreview

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Surrogate-Control", "no-store")

	for voice := range voices {
		voiceInfo, err := ttsClient.GetVoice(ctx, voices[voice].ID)
		if err != nil {
			logger("Error getting voice preview: "+err.Error(), logError)
			continue
		}
		previews = append(previews, VoicePreview{
			Name:       voices[voice].Name,
			PreviewURL: voiceInfo.PreviewURL,
		})
	}

	// sort the voices by name
	sort.Slice(previews, func(i, j int) bool {
		return previews[i].Name < previews[j].Name
	})

	data := previews
	tmpl, err := template.ParseFiles("static/voices.html")
	if err != nil {
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
