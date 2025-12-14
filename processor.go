package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents parsed input from any source (web API, Pally, etc.)
type Message struct {
	Channel         string
	Text            string
	DefaultVoice    string
	Stability       float64
	SimilarityBoost float64
	Style           float64
	PlayAlert       bool
}

// AudioSegment represents a piece of audio with voice and modifiers
type AudioSegment struct {
	Text      string
	Voice     string   // Voice ID
	VoiceName string   // Voice name for logging
	Modifiers []string // Modifiers to apply (e.g., "reverb")
	Effect    string   // Sound effect to play (empty if TTS segment)
}

// tagType represents what kind of tag was found
type tagType int

const (
	tagVoice tagType = iota
	tagModifier
	tagModifierEnd
	tagEffect
	tagUnknown
)

// ParseMessage parses a text message into audio segments
// Supports both new syntax (voicename) and legacy (v-voicename), (e-effectname)
// Note: We use () instead of [] because ElevenLabs v3 uses [] for inline audio tags
func ParseMessage(msg Message) ([]AudioSegment, error) {
	text := msg.Text
	if text == "" {
		return nil, fmt.Errorf("empty message")
	}

	// Get default voice ID
	defaultVoiceID := defaultVoiceID
	if msg.DefaultVoice != "" {
		if id, err := getVoiceID(msg.DefaultVoice); err == nil {
			defaultVoiceID = id
		}
	}

	// Parse the text into segments
	segments, err := parseTextToSegments(text, defaultVoiceID)
	if err != nil {
		return nil, err
	}

	return segments, nil
}

// parseTextToSegments handles the core parsing logic
func parseTextToSegments(text string, defaultVoiceID string) ([]AudioSegment, error) {
	var segments []AudioSegment

	// Current state
	currentVoice := defaultVoiceID
	currentVoiceName := defaultVoice
	activeModifiers := make(map[string]bool)

	// Regex to find all tags - using () instead of [] to avoid conflicts with ElevenLabs v3 audio tags
	tagRe := regexp.MustCompile(`\(([^)]+)\)`)
	matches := tagRe.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		// No tags, just return the text with default voice
		if strings.TrimSpace(text) != "" {
			segments = append(segments, AudioSegment{
				Text:      strings.TrimSpace(text),
				Voice:     currentVoice,
				VoiceName: currentVoiceName,
				Modifiers: nil,
			})
		}
		return segments, nil
	}

	// Process text between and after tags
	lastEnd := 0
	pendingText := ""

	for _, match := range matches {
		tagStart := match[0]
		tagEnd := match[1]
		tagContent := text[match[2]:match[3]] // Content inside brackets

		// Get text before this tag
		if tagStart > lastEnd {
			textBefore := strings.TrimSpace(text[lastEnd:tagStart])
			if textBefore != "" {
				pendingText += " " + textBefore
			}
		}

		// Determine tag type
		tType, name := identifyTag(tagContent)

		switch tType {
		case tagVoice:
			// If there's pending text, create a segment with current settings
			if strings.TrimSpace(pendingText) != "" {
				segments = append(segments, AudioSegment{
					Text:      strings.TrimSpace(pendingText),
					Voice:     currentVoice,
					VoiceName: currentVoiceName,
					Modifiers: getActiveModifiers(activeModifiers),
				})
				pendingText = ""
			}
			// Switch voice
			if voiceID, err := getVoiceID(name); err == nil {
				currentVoice = voiceID
				currentVoiceName = name
			} else {
				return nil, fmt.Errorf("invalid voice: %s", name)
			}

		case tagModifier:
			// If there's pending text, create segment before applying new modifier
			if strings.TrimSpace(pendingText) != "" {
				segments = append(segments, AudioSegment{
					Text:      strings.TrimSpace(pendingText),
					Voice:     currentVoice,
					VoiceName: currentVoiceName,
					Modifiers: getActiveModifiers(activeModifiers),
				})
				pendingText = ""
			}
			// Activate modifier for following text
			activeModifiers[name] = true

		case tagModifierEnd:
			// If there's pending text, create segment with current modifiers
			if strings.TrimSpace(pendingText) != "" {
				segments = append(segments, AudioSegment{
					Text:      strings.TrimSpace(pendingText),
					Voice:     currentVoice,
					VoiceName: currentVoiceName,
					Modifiers: getActiveModifiers(activeModifiers),
				})
				pendingText = ""
			}
			// Deactivate modifier
			delete(activeModifiers, name)

		case tagEffect:
			// If there's pending text, create segment first
			if strings.TrimSpace(pendingText) != "" {
				segments = append(segments, AudioSegment{
					Text:      strings.TrimSpace(pendingText),
					Voice:     currentVoice,
					VoiceName: currentVoiceName,
					Modifiers: getActiveModifiers(activeModifiers),
				})
				pendingText = ""
			}
			// Add effect segment
			segments = append(segments, AudioSegment{
				Effect: name,
			})

		case tagUnknown:
			return nil, fmt.Errorf("unknown tag: %s", tagContent)
		}

		lastEnd = tagEnd
	}

	// Handle any remaining text after the last tag
	if lastEnd < len(text) {
		remainingText := strings.TrimSpace(text[lastEnd:])
		if remainingText != "" {
			pendingText += " " + remainingText
		}
	}

	// Create final segment if there's pending text
	if strings.TrimSpace(pendingText) != "" {
		segments = append(segments, AudioSegment{
			Text:      strings.TrimSpace(pendingText),
			Voice:     currentVoice,
			VoiceName: currentVoiceName,
			Modifiers: getActiveModifiers(activeModifiers),
		})
	}

	return segments, nil
}

// identifyTag determines what kind of tag this is
func identifyTag(tagContent string) (tagType, string) {
	tagContent = strings.TrimSpace(tagContent)
	tagLower := strings.ToLower(tagContent)

	// Check for legacy syntax first
	if strings.HasPrefix(tagLower, "v-") {
		return tagVoice, tagContent[2:]
	}
	if strings.HasPrefix(tagLower, "e-") {
		return tagEffect, tagContent[2:]
	}

	// Check for modifier end tag (e.g., "reverb-end")
	if strings.HasSuffix(tagLower, "-end") {
		modifierName := tagLower[:len(tagLower)-4]
		if isModifier(modifierName) {
			return tagModifierEnd, modifierName
		}
	}

	// Check if it's a known voice
	if validVoice(tagContent) {
		return tagVoice, tagContent
	}

	// Check if it's a known modifier
	if isModifier(tagLower) {
		return tagModifier, tagLower
	}

	// Check if it's a known effect
	if _, found := getEffectSound(tagLower); found {
		return tagEffect, tagLower
	}

	return tagUnknown, tagContent
}

// getActiveModifiers converts the modifier map to a slice
func getActiveModifiers(modifiers map[string]bool) []string {
	var result []string
	for mod := range modifiers {
		result = append(result, mod)
	}
	return result
}

// ProcessAndPlay parses a message and plays the audio segments
// This is the main entry point for the unified processing pipeline
func ProcessAndPlay(msg Message) error {
	logger("Processing message through unified pipeline", logInfo, msg.Channel)

	// Parse the message into segments
	segments, err := ParseMessage(msg)
	if err != nil {
		logger("Error parsing message: "+err.Error(), logError, msg.Channel)
		return err
	}

	if len(segments) == 0 {
		logger("No segments to process", logInfo, msg.Channel)
		return nil
	}

	requestTime := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create a request entry for tracking
	trackingRequest := Request{
		Channel: msg.Channel,
		Time:    requestTime,
		Text:    msg.Text,
	}
	requests = append(requests, trackingRequest)

	// Play alert sound if requested
	if msg.PlayAlert {
		playAlertSound(msg.Channel)
	}

	// PHASE 1: Pre-generate all audio segments
	logger("Pre-generating all audio segments", logDebug, msg.Channel)
	var audioSegments [][]byte

	for _, segment := range segments {
		var audioData []byte

		if segment.Effect != "" {
			// This is an effect sound
			effectAudio, found := getEffectSound(segment.Effect)
			if !found {
				logger("Effect sound not found: "+segment.Effect, logError, msg.Channel)
				clearChannelRequests(msg.Channel)
				return fmt.Errorf("effect sound not found: %s", segment.Effect)
			}
			audioData = effectAudio
		} else if segment.Text != "" {
			// This is TTS audio
			style, err := getVoiceStyle(segment.Voice)
			if err != nil {
				style = msg.Style
			}

			ttsRequest := Request{
				Channel: msg.Channel,
				Time:    requestTime,
				Text:    segment.Text,
				Voice: TTSSettings{
					Voice:           segment.Voice,
					Stability:       msg.Stability,
					SimilarityBoost: msg.SimilarityBoost,
					Style:           style,
				},
			}

			audioData, err = generateAudio(ttsRequest)
			if err != nil {
				logger("Error generating audio: "+err.Error(), logError, msg.Channel)
				clearChannelRequests(msg.Channel)
				return err
			}

			// Apply modifiers if any
			if len(segment.Modifiers) > 0 {
				audioData = applyModifiers(audioData, segment.Modifiers, msg.Channel)
			}

			// Log data for MongoDB if enabled
			if mongoEnabled {
				data, err := createData(ttsRequest)
				if err != nil {
					logger("Error creating data: "+err.Error(), logError, msg.Channel)
				} else {
					addData(data)
				}
			}
		} else {
			continue
		}

		audioSegments = append(audioSegments, audioData)
	}

	logger(fmt.Sprintf("All %d audio segments generated, now sending", len(audioSegments)), logDebug, msg.Channel)

	// PHASE 2: Send all pre-generated audio segments
	for _, audioData := range audioSegments {
		sendTextMessage(msg.Channel, "start "+requestTime)
		time.Sleep(50 * time.Millisecond)

		sendRequest := Request{
			Channel: msg.Channel,
			Time:    requestTime,
		}
		sendAudio(sendRequest, audioData)

		// Wait for playback confirmation
		playing[requestTime] = true
		replyVerifyTicker := time.NewTicker(120 * time.Second)

		for playing[requestTime] {
			select {
			case <-replyVerifyTicker.C:
				requestName := getAudioDataName(requestTime)
				logger("No reply received for "+requestName, logInfo, msg.Channel)
				clearChannelRequests(msg.Channel)
				sendTextMessage(msg.Channel, "reload")
				return fmt.Errorf("timeout waiting for playback confirmation")
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
		replyVerifyTicker.Stop()
	}

	clearChannelRequests(msg.Channel)
	return nil
}

// playAlertSound plays the alert sound for a channel
func playAlertSound(channel string) {
	alertSound, alertExists := getAlertSound(channel)
	if !alertExists {
		return
	}

	alertSoundBytes, err := io.ReadAll(alertSound)
	if err != nil {
		logger("Error reading alert sound: "+err.Error(), logError, channel)
		return
	}

	waitTime, err := getAudioLengthFile(alertSound.Name())
	if err != nil {
		logger("Error getting alert length: "+err.Error(), logError, channel)
		waitTime = 5
	}

	for client, clientChannel := range clients {
		if clientChannel == channel {
			clientName := getClientName(fmt.Sprintf("%p", client))
			err := client.WriteMessage(websocket.BinaryMessage, alertSoundBytes)
			if err != nil {
				logger("Error sending alert sound to "+clientName+": "+err.Error(), logError, channel)
				client.Close()
				connMutex.Lock()
				delete(clients, client)
				connMutex.Unlock()
			} else {
				logger("Alert sound sent to "+clientName, logInfo, channel)
			}
		}
	}

	time.Sleep(time.Duration(waitTime) * time.Second)
}

// isModifier checks if a string is a known modifier name
func isModifier(name string) bool {
	modifiers := []string{"reverb"} // Add more modifiers here
	for _, m := range modifiers {
		if strings.EqualFold(m, name) {
			return true
		}
	}
	return false
}
