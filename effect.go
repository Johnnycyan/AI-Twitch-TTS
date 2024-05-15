package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	effectFolder = "effects"
)

func getEffectSound(effect string) ([]byte, bool) {
	var effectSounds []string
	//check if alert folder exists and if so get all .mp3 files in it
	if _, err := os.Stat(effectFolder); err == nil {
		files, err := os.ReadDir(effectFolder)
		if err != nil {
			logger("Error reading alert folder: "+err.Error(), logError)
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

	// check if there are any alert sounds in the folder
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
		logger("Error opening alert sound: "+err.Error(), logError)
		return nil, false
	}

	effectSoundBytes, err := io.ReadAll(effectSound)

	return effectSoundBytes, true
}
