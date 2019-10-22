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
	Data         interface{}
}

const (
	// RIFFHeaderLen is the length of RIFF header in bytes.
	RIFFHeaderLen = 12
	// FmtChunkLen is the length of fmt chunk in bytes.
	FmtChunkLen = 24
)

// SineWave16 generates the 16 bit monaural sine wave with given parameters.
func SineWave(channel int, stereo bool, frequency, amplitude float64, samplingRate int) Waveform {
	waveform := Waveform{
		SamplingRate: samplingRate,
		Channel:      channel,
		Stereo:       false,
	}

	switch channel {
	case 16:
		soundData := make([]int16, 0, samplingRate)
		for i := 0; i < samplingRate; i++ {
			value := amplitude * math.Sin(2*math.Pi*frequency*float64(i)/float64(samplingRate))
			value *= 32767.0
			v := int16(value)

			// clipping to adjust for the range of 16 bit integer
			if v > 32767 {
				fmt.Println(v)
				v = 32767
			} else if v < -32768 {
				fmt.Println(v)
				v = -32768
			}
			soundData = append(soundData, v)
		}
		waveform.Data = soundData
	default:
		log.Println("unsupported parameters")
	}
	return waveform
}

// Write writes given audio data to wave file.
func Write(filename string, waveform Waveform) error {
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
		BitsPerSample: 16,
	}
	fmtChunk.BlockSize = fmtChunk.BitsPerSample * fmtChunk.Channel / 8
	fmtChunk.BytesPerSec = uint32(fmtChunk.BlockSize) * fmtChunk.SamplesPerSec
	dataChunk := DataChunk{
		ChunkID: [4]byte{'d', 'a', 't', 'a'},
	}

	var b bytes.Buffer
	switch waveform.Channel {
	case 16:
		soundData := waveform.Data.([]int16)
		for _, d := range soundData {
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

// FadeIn applies fade-in to given soundData.
// duration is in millisecond.
func FadeIn(waveform Waveform, duration int) Waveform {
	fadePeriod := waveform.SamplingRate / 1000 * duration

	ret := Waveform{
		SamplingRate: waveform.SamplingRate,
		Channel:      waveform.Channel,
	}

	switch waveform.Channel {
	case 16:
		soundData := waveform.Data.([]int16)
		for i := 0; i < fadePeriod; i++ {
			amplitudeRate := float64(i) / float64(fadePeriod)
			soundData[i] = int16(float64(soundData[i]) * amplitudeRate)
		}
		ret.Data = soundData
	default:
		log.Println("given Waveform has not supported")
		return waveform
	}
	return ret
}

// FadeOut applies fade-out to given soundData.
// duration is in millisecond.
func FadeOut(waveform Waveform, duration int) Waveform {
	fadePeriod := waveform.SamplingRate / 1000 * duration

	ret := Waveform{
		SamplingRate: waveform.SamplingRate,
		Channel:      waveform.Channel,
	}

	switch waveform.Channel {
	case 16:
		soundData := waveform.Data.([]int16)
		startIdx := len(soundData) - fadePeriod
		for i := 0; i < fadePeriod; i++ {
			amplitudeRate := (float64(fadePeriod) - float64(i)) / float64(fadePeriod)
			soundData[startIdx+i] = int16(float64(soundData[startIdx+i]) * amplitudeRate)
		}
		ret.Data = soundData
	default:
		log.Println("given Waveform has not supported")
		return waveform
	}

	return ret
}
