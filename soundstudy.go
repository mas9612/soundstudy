package main

import (
	"log"

	"github.com/mas9612/soundstudy/wave"
)

func main() {
	// generate sine wave
	frequency := 440.0
	samplingRate := 44100

	waveform := wave.SineWave(16, false, frequency, 1.0, samplingRate)
	for i := 2; i <= 44; i++ {
		wf := wave.SineWave(16, false, frequency*float64(i), 1.0/float64(i), samplingRate)

		var err error
		waveform, err = wave.Add(waveform, wf)
		if err != nil {
			log.Fatal(err)
		}
	}

	waveform = wave.Gain(waveform, 0.5)

	waveform = wave.FadeIn(waveform, 10)
	waveform = wave.FadeOut(waveform, 10)

	if err := wave.WriteWaveData("wavedata", waveform); err != nil {
		log.Fatal(err)
	}

	if err := wave.Write("test.wav", waveform); err != nil {
		log.Fatal(err)
	}
}
