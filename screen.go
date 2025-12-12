package main

import (
	"io"
)

type Screen struct {
	Width    int
	Height   int
	rowBytes int
	buffer   []byte
}

func NewScreen(w, h int) Screen {
	rowBytes := (w + 7) / 8
	buf := make([]byte, rowBytes*h)

	return Screen{Width: w, Height: h, rowBytes: rowBytes, buffer: buf}
}

func (s *Screen) Set(x, y int, on bool) {
	if x < 0 || y < 0 || x >= s.Width || y >= s.Height {
		return
	}

	byteIndex := y*s.rowBytes + (x / 8)
	bitPos := 7 - (x % 8)

	if on {
		s.buffer[byteIndex] |= 1 << bitPos
	} else {
		s.buffer[byteIndex] &^= 1 << bitPos
	}
}

func (s *Screen) Draw(x, y int, data []byte) {
	for row := range len(data) {
		for bit := range 8 {
			on := (data[row]>>(7-bit))&1 == 1
			s.Set(x+bit, y+row, on)
		}
	}
}

func (s *Screen) Render(w io.Writer) error {
	line := make([]byte, s.Width)

	for y := range s.Height {
		for x := range s.Width {
			byteIndex := y*s.rowBytes + (x / 8)
			bitPos := 7 - (x % 8)
			bit := (s.buffer[byteIndex] >> bitPos & 1)

			if bit == 1 {
				line[x] = '#'
			} else {
				line[x] = ' '
			}
		}
		if _, err := w.Write(line); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

func (s *Screen) Clear() {
	for i := range s.buffer {
		s.buffer[i] = 0
	}
}
