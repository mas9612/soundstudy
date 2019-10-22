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

	waveform := wave.SineWave(16, false, frequency, amplitude, samplingRate)
	waveform = wave.FadeIn(waveform, 10)
	waveform = wave.FadeOut(waveform, 10)

	if err := wave.Write("test.wav", waveform); err != nil {
		log.Fatal(err)
	}
}
