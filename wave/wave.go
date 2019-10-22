package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
)

// RIFFHeader represents the header of RIFF chunk.
type RIFFHeader struct {
	ChunkID    [4]byte
	ChunkSize  uint32
	FormatType [4]byte
}

// FmtChunk represents the fmt chunk of WAVE file.
type FmtChunk struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	FormatType    uint16
	Channel       uint16
	SamplesPerSec uint32
	BytesPerSec   uint32
	BlockSize     uint16
	BitsPerSample uint16
}

// DataChunk represents the data chunk of WAVE file.
type DataChunk struct {
	ChunkID   [4]byte
	ChunkSize uint32
	Data      []byte
}

type Waveform struct {
	SamplingRate int
	Channel      int
	Stereo       bool
	Data         []float64
	max          float64
}

const (
	// RIFFHeaderLen is the length of RIFF header in bytes.
	RIFFHeaderLen = 12
	// FmtChunkLen is the length of fmt chunk in bytes.
	FmtChunkLen = 24
)

// SineWave16 generates the 16 bit monaural sine wave with given parameters.
func SineWave(channel int, stereo bool, frequency, amplitude float64, samplingRate int) *Waveform {
	waveform := &Waveform{
		SamplingRate: samplingRate,
		Channel:      channel,
		Stereo:       false,
	}

	switch channel {
	case 16:
		soundData := make([]float64, 0, samplingRate)
		for i := 0; i < samplingRate; i++ {
			value := amplitude * math.Sin(2*math.Pi*frequency*float64(i)/float64(samplingRate))
			if value > waveform.max {
				waveform.max = value
			}
			soundData = append(soundData, value)
		}
		waveform.Data = soundData
	default:
		log.Println("unsupported parameters")
	}
	return waveform
}

func normalize(waveform *Waveform) (interface{}, error) {
	switch waveform.Channel {
	case 8:
		data := make([]uint8, len(waveform.Data))
		return data, nil
	case 16:
		data := make([]int16, len(waveform.Data))

		for i, d := range waveform.Data {
			value := d / waveform.max * 32767.0
			if value > 32767 {
				value = 32767
			} else if value < -32768 {
				value = -32768
			}

			data[i] = int16(value)
		}

		return data, nil
	default:
		return nil, fmt.Errorf("normalize: invalid waveform data")
	}
}

// Write writes given audio data to wave file.
func Write(filename string, waveform *Waveform) error {
	hdr := RIFFHeader{
		ChunkID:    [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:  RIFFHeaderLen + FmtChunkLen,
		FormatType: [4]byte{'W', 'A', 'V', 'E'},
	}
	fmtChunk := FmtChunk{
		ChunkID:       [4]byte{'f', 'm', 't', ' '},
		ChunkSize:     FmtChunkLen - 8, // FmtChunkLen - len(ChunkID) - len(ChunkSize)
		FormatType:    1,
		Channel:       1,
		SamplesPerSec: uint32(waveform.SamplingRate),
		BitsPerSample: uint16(waveform.Channel),
	}
	fmtChunk.BlockSize = fmtChunk.BitsPerSample * fmtChunk.Channel / 8
	fmtChunk.BytesPerSec = uint32(fmtChunk.BlockSize) * fmtChunk.SamplesPerSec
	dataChunk := DataChunk{
		ChunkID: [4]byte{'d', 'a', 't', 'a'},
	}

	normalizedSoundData, err := normalize(waveform)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	switch waveform.Channel {
	case 16:
		for _, d := range normalizedSoundData.([]int16) {
			if err := binary.Write(&b, binary.LittleEndian, d); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("given waveform has not supported")
	}
	dataChunk.ChunkSize = uint32(b.Len() * int(fmtChunk.Channel))
	dataChunk.Data = make([]byte, dataChunk.ChunkSize)
	copy(dataChunk.Data, b.Bytes())

	buf := make([]byte, hdr.ChunkSize+8+dataChunk.ChunkSize)

	copy(buf, hdr.ChunkID[:])
	binary.LittleEndian.PutUint32(buf[4:], hdr.ChunkSize)
	copy(buf[8:], hdr.FormatType[:])

	copy(buf[12:], fmtChunk.ChunkID[:])
	binary.LittleEndian.PutUint32(buf[16:], fmtChunk.ChunkSize)
	binary.LittleEndian.PutUint16(buf[20:], fmtChunk.FormatType)
	binary.LittleEndian.PutUint16(buf[22:], fmtChunk.Channel)
	binary.LittleEndian.PutUint32(buf[24:], fmtChunk.SamplesPerSec)
	binary.LittleEndian.PutUint32(buf[28:], fmtChunk.BytesPerSec)
	binary.LittleEndian.PutUint16(buf[32:], fmtChunk.BlockSize)
	binary.LittleEndian.PutUint16(buf[34:], fmtChunk.BitsPerSample)

	copy(buf[36:], dataChunk.ChunkID[:])
	binary.LittleEndian.PutUint32(buf[40:], dataChunk.ChunkSize)
	copy(buf[44:], dataChunk.Data)

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := file.Write(buf)
	if err != nil {
		return err
	}
	if n < len(buf) {
		return fmt.Errorf("short write")
	}

	return nil
}

// WriteWaveData writes wave data to given filename to plot wave with gnuplot.
func WriteWaveData(filename string, waveform *Waveform) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for i, d := range waveform.Data {
		fmt.Fprintf(file, "%d %f\n", i+1, d)
	}

	return nil
}

// FadeIn applies fade-in to given soundData.
// duration is in millisecond.
func FadeIn(waveform *Waveform, duration int) *Waveform {
	fadePeriod := waveform.SamplingRate / 1000 * duration

	ret := &Waveform{
		SamplingRate: waveform.SamplingRate,
		Channel:      waveform.Channel,
		Stereo:       waveform.Stereo,
		Data:         waveform.Data,
		max:          waveform.max,
	}

	for i := 0; i < fadePeriod; i++ {
		amplitudeRate := float64(i) / float64(fadePeriod)
		ret.Data[i] = waveform.Data[i] * amplitudeRate
	}

	return ret
}

// FadeOut applies fade-out to given soundData.
// duration is in millisecond.
func FadeOut(waveform *Waveform, duration int) *Waveform {
	fadePeriod := waveform.SamplingRate / 1000 * duration

	ret := &Waveform{
		SamplingRate: waveform.SamplingRate,
		Channel:      waveform.Channel,
		Stereo:       waveform.Stereo,
		Data:         waveform.Data,
		max:          waveform.max,
	}

	startIdx := len(waveform.Data) - fadePeriod
	for i := 0; i < fadePeriod; i++ {
		amplitudeRate := (float64(fadePeriod) - float64(i)) / float64(fadePeriod)
		ret.Data[startIdx+i] = waveform.Data[startIdx+i] * amplitudeRate
	}

	return ret
}

// Add adds two wave and returns the new waveform data.
func Add(wave1, wave2 *Waveform) (*Waveform, error) {
	if wave1.Channel != wave2.Channel {
		return nil, fmt.Errorf("wave1 and wave2 must have same channel bit")
	}
	if wave1.SamplingRate != wave2.SamplingRate {
		return nil, fmt.Errorf("wave1 and wave2 must have same sampling rate")
	}
	if wave1.Stereo != wave2.Stereo {
		return nil, fmt.Errorf("wave1 and wave2 must be both stereo or both monaural")
	}

	waveform := &Waveform{
		Channel:      wave1.Channel,
		SamplingRate: wave1.SamplingRate,
		Stereo:       wave1.Stereo,
	}
	switch waveform.Channel {
	case 16:
		newSoundLen := intMax(len(wave1.Data), len(wave1.Data))
		waveform.Data = make([]float64, newSoundLen)
		for i := 0; i < newSoundLen; i++ {
			waveform.Data[i] = wave1.Data[i] + wave2.Data[i]
			if waveform.Data[i] > waveform.max {
				waveform.max = waveform.Data[i]
			}
		}
	default:
		return nil, fmt.Errorf("invalid channel '%d'", waveform.Channel)
	}

	return waveform, nil
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
