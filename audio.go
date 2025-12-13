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
	writer     io.Writer
	playing    bool
	stpCh      chan struct{}
}

func NewAudio(sr, ch, bit int, w io.Writer) *Audio {
	return &Audio{
		SampleRate: sr,
		Channels:   ch,
		BitDepth:   bit,
		writer:     w,
		stpCh:      make(chan struct{}),
	}
}

func (a *Audio) On() {
	if a.playing {
		return
	}

	a.playing = true
	a.stpCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-a.stpCh:
				return
			default:
				samples := a.GenerateSquareWave(440, 0.05, 3000)
				_ = a.WriteSamples(samples)
			}
		}
	}()
}

func (a *Audio) Off() {
	if !a.playing {
		return
	}

	close(a.stpCh)
	a.playing = false
}

func (a *Audio) WriteSamples(samples []int16) error {
	if a.BitDepth != 16 {
		return errors.New("only 16bits sample audio")
	}

	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}

	_, err := a.writer.Write(buf)
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
