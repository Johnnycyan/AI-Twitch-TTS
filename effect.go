package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
			logger("Error reading effects folder: "+err.Error(), logError, "Universal")
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

func getEffectSound(effect string) ([]byte, bool) {
	logger("Getting effect sound for: "+effect, logDebug, "Universal")
	var effectSounds []string
	//check if effects folder exists and if so get all .mp3 files in it
	if _, err := os.Stat(effectFolder); err == nil {
		files, err := os.ReadDir(effectFolder)
		if err != nil {
			logger("Error reading effects folder: "+err.Error(), logError, "Universal")
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
		logger("No effect sounds found in effect folder: "+effectFolder, logDebug, "Universal")
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
	if effectSelected == "" {
		logger("No effect sound found for: "+effect, logDebug, "Universal")
		return nil, false
	}
	effectSound, err := os.Open(fmt.Sprintf("%s/%s", effectFolder, effectSelected))
	if err != nil {
		logger("Error opening effect sound: "+err.Error(), logError, "Universal")
		return nil, false
	}

	effectSoundBytes, err := io.ReadAll(effectSound)
	if err != nil {
		logger("Error reading effect sound: "+err.Error(), logError, "Universal")
		return nil, false
	}

	return effectSoundBytes, true
}
