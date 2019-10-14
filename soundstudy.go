package main

import (
	"bytes"
	"encoding/binary"
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

const (
	// RIFFHeaderLen is the length of RIFF header in bytes.
	RIFFHeaderLen = 12
	// FmtChunkLen is the length of fmt chunk in bytes.
	FmtChunkLen = 24
)

func main() {
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
		SamplesPerSec: 44100,
		BitsPerSample: 16,
	}
	fmtChunk.BlockSize = fmtChunk.BitsPerSample * fmtChunk.Channel / 8
	fmtChunk.BytesPerSec = uint32(fmtChunk.BlockSize) * fmtChunk.SamplesPerSec
	dataChunk := DataChunk{
		ChunkID: [4]byte{'d', 'a', 't', 'a'},
	}

	// generate sine wave
	samplingFreq := 44100
	amplitude := 1
	frequency := 440
	var b bytes.Buffer
	for i := 0; i < samplingFreq; i++ {
		value := float64(amplitude) * math.Sin(2*math.Pi*float64(frequency)*float64(i)/float64(samplingFreq))
		value = value / float64(amplitude) * 32767.0
		v := int16(value)

		// clipping to adjust for the range of 16 bit integer
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		binary.Write(&b, binary.LittleEndian, v)
	}
	dataChunk.ChunkSize = uint32(b.Len() * int(fmtChunk.Channel))
	dataChunk.Data = make([]byte, dataChunk.ChunkSize)
	copy(dataChunk.Data, b.Bytes())

	// create wave file
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

	file, err := os.OpenFile("test.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	n, err := file.Write(buf)
	if err != nil {
		log.Fatal(err)
	}
	if n < len(buf) {
		log.Fatal("short write")
	}
}
