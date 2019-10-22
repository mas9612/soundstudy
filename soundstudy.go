package main

import (
	"log"

	"github.com/mas9612/soundstudy/wave"
)

func main() {
	// generate sine wave
	frequency := 440.0
	amplitude := 1.0
	samplingRate := 44100

	soundData := wave.SineWave16(frequency, amplitude, samplingRate)
	soundData = wave.FadeIn(soundData, samplingRate, 10)
	soundData = wave.FadeOut(soundData, samplingRate, 10)

	if err := wave.Write16bitMonaural("test.wav", soundData, samplingRate); err != nil {
		log.Fatal(err)
	}
}
