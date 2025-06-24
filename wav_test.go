package tag

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func createTestWAV(sampleRate uint32, channels uint16, bitsPerSample uint16, durationSeconds float64) []byte {
	var buf bytes.Buffer

	// Calculate data size
	bytesPerSample := (bitsPerSample + 7) / 8
	totalSamples := uint32(float64(sampleRate) * durationSeconds)
	dataSize := totalSamples * uint32(channels) * uint32(bytesPerSample)

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize)) // file size - 8
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16)) // fmt chunk size
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // audio format (PCM)
	binary.Write(&buf, binary.LittleEndian, channels)
	binary.Write(&buf, binary.LittleEndian, sampleRate)
	binary.Write(&buf, binary.LittleEndian, sampleRate*uint32(channels)*uint32(bytesPerSample)) // byte rate
	binary.Write(&buf, binary.LittleEndian, uint16(channels)*uint16(bytesPerSample))            // block align
	binary.Write(&buf, binary.LittleEndian, bitsPerSample)

	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, dataSize)

	// Write dummy audio data (silence)
	silence := make([]byte, dataSize)
	buf.Write(silence)

	return buf.Bytes()
}

func TestWAVDuration(t *testing.T) {
	tests := []struct {
		name          string
		sampleRate    uint32
		channels      uint16
		bitsPerSample uint16
		durationSecs  float64
		expectedDur   time.Duration
	}{
		{
			name:          "44.1kHz 16-bit stereo 5 seconds",
			sampleRate:    44100,
			channels:      2,
			bitsPerSample: 16,
			durationSecs:  5.0,
			expectedDur:   5 * time.Second,
		},
		{
			name:          "48kHz 24-bit mono 3 seconds",
			sampleRate:    48000,
			channels:      1,
			bitsPerSample: 24,
			durationSecs:  3.0,
			expectedDur:   3 * time.Second,
		},
		{
			name:          "22.05kHz 8-bit stereo 2.5 seconds",
			sampleRate:    22050,
			channels:      2,
			bitsPerSample: 8,
			durationSecs:  2.5,
			expectedDur:   2500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wavData := createTestWAV(tt.sampleRate, tt.channels, tt.bitsPerSample, tt.durationSecs)
			reader := bytes.NewReader(wavData)

			meta, err := ReadWAVMeta(reader)
			if err != nil {
				t.Fatalf("ReadWAVMeta failed: %v", err)
			}

			if meta.FileType() != WAV {
				t.Errorf("Expected file type WAV, got %v", meta.FileType())
			}

			duration := meta.Duration()

			// Allow for small rounding differences (within 1ms)
			diff := duration - tt.expectedDur
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Millisecond {
				t.Errorf("Expected duration %v, got %v (diff: %v)", tt.expectedDur, duration, diff)
			}

			// Test Raw() method returns expected values
			raw := meta.Raw()
			if raw["sample_rate"] != tt.sampleRate {
				t.Errorf("Expected sample rate %d, got %v", tt.sampleRate, raw["sample_rate"])
			}
			if raw["channels"] != tt.channels {
				t.Errorf("Expected channels %d, got %v", tt.channels, raw["channels"])
			}
			if raw["bits_per_sample"] != tt.bitsPerSample {
				t.Errorf("Expected bits per sample %d, got %v", tt.bitsPerSample, raw["bits_per_sample"])
			}
		})
	}
}

func TestWAVMetadataInterface(t *testing.T) {
	wavData := createTestWAV(44100, 2, 16, 1.0)
	reader := bytes.NewReader(wavData)

	meta, err := ReadWAVMeta(reader)
	if err != nil {
		t.Fatalf("ReadWAVMeta failed: %v", err)
	}

	// Test that all metadata interface methods work (should return empty/zero values for WAV)
	if meta.Title() != "" {
		t.Errorf("Expected empty title, got %q", meta.Title())
	}
	if meta.Album() != "" {
		t.Errorf("Expected empty album, got %q", meta.Album())
	}
	if meta.Artist() != "" {
		t.Errorf("Expected empty artist, got %q", meta.Artist())
	}
	if meta.Year() != 0 {
		t.Errorf("Expected year 0, got %d", meta.Year())
	}
	if track, total := meta.Track(); track != 0 || total != 0 {
		t.Errorf("Expected track (0, 0), got (%d, %d)", track, total)
	}
	if meta.Picture() != nil {
		t.Errorf("Expected nil picture, got %v", meta.Picture())
	}
}
