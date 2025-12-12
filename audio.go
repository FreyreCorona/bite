package main

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

type Audio struct {
	SampleRate int
	Channels   int
	BitDepth   int
}

func NewAudio(sr, ch, bit int) *Audio {
	return &Audio{sr, ch, bit}
}

func (a *Audio) WriteSamples(w io.Writer, samples []int16) error {
	if a.BitDepth != 16 {
		return errors.New("only 16bits sample audio")
	}

	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}

	_, err := w.Write(buf)
	return err
}

func (a *Audio) GenerateSquareWave(freq float64, durationSec float64, amplitude int16) []int16 {
	total := int(float64(a.SampleRate) * durationSec)
	samples := make([]int16, total)

	period := float64(a.SampleRate) / freq

	for i := range total {
		if math.Mod(float64(i), period) < period/2 {
			samples[i] = amplitude
		} else {
			samples[i] = -amplitude
		}
	}

	return samples
}

func (a *Audio) GenerateSine(freq float64, durationSec float64, amplitude int16) []int16 {
	total := int(float64(a.SampleRate) * durationSec)
	samples := make([]int16, total)

	for i := range total {
		t := float64(i) / float64(a.SampleRate)
		samples[i] = int16(float64(amplitude) * math.Sin(2*math.Pi*freq*t))
	}

	return samples
}
