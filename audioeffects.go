package main

import (
	"regexp"
	"sort"
	"strings"
)

// AudioEffects holds the ElevenLabs v4 audio_effects parameters.
// Values are set by <> tags in message text and apply to following text
// until replaced or cleared.
type AudioEffects struct {
	FilterPresetID    string
	EnvironmentID     string
	BackgroundNoiseID string
	Distance          float64
}

// IsDefault returns true when no audio effect is active
func (ae *AudioEffects) IsDefault() bool {
	if ae == nil {
		return true
	}
	return ae.FilterPresetID == "" && ae.EnvironmentID == "" && ae.BackgroundNoiseID == "" && ae.Distance == 0
}

// snapshot returns a copy of the current effects, or nil when none are active
func (ae *AudioEffects) snapshot() *AudioEffects {
	if ae.IsDefault() {
		return nil
	}
	cp := *ae
	return &cp
}

// audioEffectTags maps a <tag> name to the state change it applies.
// Set tags replace the current value of their slot; off tags clear it.
var audioEffectTags = map[string]func(*AudioEffects){
	// filter_preset_id
	"old_radio":         setFilterPreset("old_radio"),
	"robot":             setFilterPreset("robot"),
	"cheap_microphone":  setFilterPreset("cheap_microphone"),
	"phone":             setFilterPreset("phone"),
	"low_quality_phone": setFilterPreset("low_quality_phone"),
	"bright_phone":      setFilterPreset("bright_phone"),

	// environment_id
	"small_room": setEnvironment("small_room"),
	"big_room":   setEnvironment("big_room"),
	"hall":       setEnvironment("hall"),
	"tunnel":     setEnvironment("tunnel"),
	"street":     setEnvironment("street"),
	"valley":     setEnvironment("valley"),
	"forest":     setEnvironment("forest"),

	// background_noise_id
	"call_center": setBackgroundNoise("call_center"),
	"cafe":        setBackgroundNoise("cafe"),
	"city":        setBackgroundNoise("city"),
	"keyboard":    setBackgroundNoise("keyboard"),

	// distance (0 = none, 1 = far)
	"near":   setDistance(0.25),
	"medium": setDistance(0.6),
	"far":    setDistance(1.0),

	// off tags
	"filter-off":      setFilterPreset(""),
	"environment-off": setEnvironment(""),
	"noise-off":       setBackgroundNoise(""),
	"distance-off":    setDistance(0),
	"effects-off":     clearAudioEffects,
}

func setFilterPreset(v string) func(*AudioEffects) {
	return func(ae *AudioEffects) { ae.FilterPresetID = v }
}

func setEnvironment(v string) func(*AudioEffects) {
	return func(ae *AudioEffects) { ae.EnvironmentID = v }
}

func setBackgroundNoise(v string) func(*AudioEffects) {
	return func(ae *AudioEffects) { ae.BackgroundNoiseID = v }
}

func setDistance(v float64) func(*AudioEffects) {
	return func(ae *AudioEffects) { ae.Distance = v }
}

func clearAudioEffects(ae *AudioEffects) {
	*ae = AudioEffects{}
}

// orNull converts an empty string to nil so it serializes as JSON null
func orNull(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// identifyAudioEffectTag reports whether the content of a <> tag is a known
// audio effect tag and returns the function that applies it to the state
func identifyAudioEffectTag(content string) (func(*AudioEffects), bool) {
	apply, found := audioEffectTags[strings.ToLower(strings.TrimSpace(content))]
	return apply, found
}

// tagRegex is the combined tag regex: group 1 captures () tag content,
// group 2 captures <> audio effect tag content. Only known audio effect names
// are matched inside <> so unknown angle brackets stay literal text.
var tagRegex = buildTagRegex()

func buildTagRegex() *regexp.Regexp {
	names := make([]string, 0, len(audioEffectTags))
	for name := range audioEffectTags {
		names = append(names, regexp.QuoteMeta(name))
	}
	sort.Strings(names)
	return regexp.MustCompile(`\(([^)]+)\)|<((?i:` + strings.Join(names, "|") + `))>`)
}
