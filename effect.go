package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
)

var (
	effectFolder = "effects"
)

type FileListData struct {
	Files []string
}

func listEffects(w http.ResponseWriter, _ *http.Request) {
	var responseString string
	//check if effect folder exists and if so get all .mp3 files in it
	if _, err := os.Stat(effectFolder); err == nil {
		files, err := os.ReadDir(effectFolder)
		if err != nil {
			logger("Error reading effects folder: "+err.Error(), logError)
			w.Write([]byte("Error reading effect folder: " + err.Error()))
			return
		} else {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".mp3") {
					responseString += strings.TrimSuffix(file.Name(), ".mp3") + ", "
				}
			}
		}
	} else {
		w.Write([]byte("Effects folder does not exist"))
		return
	}

	responseString = strings.TrimSuffix(responseString, ", ")

	w.Write([]byte(responseString))
}

func effectsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the URL path is exactly "/effects"
	if r.URL.Path != "/effects" {
		http.NotFound(w, r)
		return
	}

	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Surrogate-Control", "no-store")

	dir, err := os.Open(effectFolder)
	if err != nil {
		http.Error(w, "Unable to open effects directory", http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		http.Error(w, "Unable to read directory", http.StatusInternalServerError)
		return
	}

	var mp3Files []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mp3") {
			// remove the .mp3 extension from the file name
			mp3Files = append(mp3Files, strings.TrimSuffix(file.Name(), ".mp3"))
		}
	}

	slices.Sort(mp3Files)

	data := FileListData{Files: mp3Files}
	tmpl, err := template.ParseFiles("effects.html")
	if err != nil {
		http.Error(w, "Unable to parse template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func getEffectSound(effect string) ([]byte, bool) {
	var effectSounds []string
	//check if effects folder exists and if so get all .mp3 files in it
	if _, err := os.Stat(effectFolder); err == nil {
		files, err := os.ReadDir(effectFolder)
		if err != nil {
			logger("Error reading effects folder: "+err.Error(), logError)
			return nil, false
		} else {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".mp3") {
					effectSounds = append(effectSounds, file.Name())
				}
			}
		}
	} else {
		return nil, false
	}

	// check if there are any effect sounds in the folder
	if len(effectSounds) == 0 {
		logger("No effect sounds found in effect folder: "+effectFolder, logDebug)
		return nil, false
	}

	// get the effect sound from the list of effect sounds if the name of the effect is contained in the file name
	var effectSelected string
	for _, sound := range effectSounds {
		if strings.Contains(sound, effect) {
			effectSelected = sound
			break
		}
	}
	effectSound, err := os.Open(fmt.Sprintf("%s/%s", effectFolder, effectSelected))
	if err != nil {
		logger("Error opening effect sound: "+err.Error(), logError)
		return nil, false
	}

	effectSoundBytes, err := io.ReadAll(effectSound)

	return effectSoundBytes, true
}
