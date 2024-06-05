package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

var (
	alertFolder = "alerts"
)

func getAlertSound(channel string) (*os.File, bool) {
	var alertSounds []string
	channelAlertsFolder := fmt.Sprintf("%s/%s", alertFolder, channel)
	//check if alert folder exists and if so get all .mp3 files in it
	if _, err := os.Stat(channelAlertsFolder); err == nil {
		files, err := os.ReadDir(channelAlertsFolder)
		if err != nil {
			logger("Error reading alert folder: "+err.Error(), logError, channel)
			return nil, false
		} else {
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".mp3") {
					alertSounds = append(alertSounds, file.Name())
				}
			}
		}
	} else {
		return nil, false
	}

	// check if there are any alert sounds in the folder
	if len(alertSounds) == 0 {
		logger("No alert sounds found in alert folder: "+channelAlertsFolder, logDebug, channel)
		return nil, false
	}

	// get a random alert sound from the list of alert sounds
	randomAlertSound := alertSounds[rand.Intn(len(alertSounds))]
	logger("Random alert sound selected: "+randomAlertSound, logDebug, channel)
	alertSound, err := os.Open(fmt.Sprintf("%s/%s", channelAlertsFolder, randomAlertSound))
	if err != nil {
		logger("Error opening alert sound: "+err.Error(), logError, channel)
		return nil, false
	}

	return alertSound, true
}
