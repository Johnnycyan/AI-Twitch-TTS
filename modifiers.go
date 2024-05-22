package main

import (
	"log"
	"os"
	"os/exec"
)

func loadAudioDataFromFile(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("File not found")
		log.Fatal(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Println("Failed to get file info")
		log.Fatal(err)
	}

	data := make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		log.Println("Failed to read file")
		log.Fatal(err)
	}

	return data
}

func saveAudioDataToFile(filename string, data []byte) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Failed to create file")
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		log.Println("Failed to write to file")
		log.Fatal(err)
	}
}

func deleteAudioFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		log.Println("Failed to delete file: " + filename)
		log.Fatal(err)
	}
}

func addReverbToAudio(channel string) {
	// ffmpeg -i input.mp3 -i "reverb.wav" -filter_complex "[0:a]apad=pad_dur=2[dry];[0:a]apad=pad_dur=2,afir=dry=10:wet=10[wet];[dry][wet]amix=weights='0.8 0.2'" -b:a 320k output.mp3
	cmd := exec.Command("ffmpeg", "-i", "input-"+channel+".mp3", "-i", "reverb.wav", "-filter_complex", "[0:a]apad=pad_dur=2,aformat=channel_layouts=stereo[dry];[0:a]apad=pad_dur=2,aformat=channel_layouts=stereo,afir=dry=10:wet=10[wet];[dry][wet]amix=weights='0.8 0.2'", "-b:a", "320k", "output-"+channel+".mp3")
	err := cmd.Run()
	if err != nil {
		log.Println("Failed to add reverb to audio")
		log.Fatal(err)
	}
}

func reverb(data []byte, channel string) []byte {
	// Save audio data to file
	saveAudioDataToFile("input-"+channel+".mp3", data)

	// Add reverb to audio
	addReverbToAudio(channel)

	// Delete the input file
	deleteAudioFile("input-" + channel + ".mp3")

	// Load the reverb data from the output file
	reverbData := loadAudioDataFromFile("output-" + channel + ".mp3")

	// Delete the output file
	deleteAudioFile("output-" + channel + ".mp3")

	return reverbData
}
