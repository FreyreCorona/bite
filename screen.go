package main

import (
	"io"
)

type Screen struct {
	Width    int
	Height   int
	rowBytes int
	scale    int
	buffer   []byte
}

func NewScreen(w, h, scale int) Screen {
	rowBytes := (w + 7) / 8
	buf := make([]byte, rowBytes*h)

	if scale < 1 {
		scale = 1
	}

	return Screen{
		Width:    w,
		Height:   h,
		rowBytes: rowBytes,
		scale:    scale,
		buffer:   buf,
	}
}

func (s *Screen) Set(x, y int, on bool) {
	if x < 0 || y < 0 || x >= s.Width || y >= s.Height {
		return
	}

	i := y*s.rowBytes + (x / 8)
	pos := 7 - (x % 8)

	if on {
		s.buffer[i] |= 1 << pos
		return
	}

	s.buffer[i] &^= 1 << pos
}

func (s *Screen) Get(x, y int) byte {
	if x < 0 || y < 0 || x >= s.Width || y >= s.Height {
		return 0
	}

	i := y*s.rowBytes + (x / 8)
	if i < 0 || i >= len(s.buffer) {
		return 0
	}

	return s.buffer[i]
}

func (s *Screen) Draw(x, y int, data []byte) {
	n := len(s.buffer)

	for row := range len(data) {
		dy := y + row
		if dy < 0 || dy >= s.Height {
			continue
		}

		sprByte := data[row]
		i := dy*s.rowBytes + (x / 8)

		shift := x % 8

		if shift == 0 {
			if i >= 0 && i < n {
				s.buffer[i] = s.buffer[i] ^ sprByte
			}
		} else {
			if i < 0 || i >= n {
				continue
			}

			l := sprByte >> shift
			r := sprByte << (8 - shift)

			s.buffer[i] = s.buffer[i] ^ l

			if i+1 < n {
				s.buffer[i+1] = s.buffer[i+1] ^ r
			}
		}
	}
}

func (s *Screen) Render(w io.Writer) error {
	on := '█'
	off := ' '

	line := make([]rune, s.Width*s.scale)

	for y := range s.Height {
		for x := range s.Width {
			i := y*s.rowBytes + (x / 8)
			pos := 7 - (x % 8)
			bit := (s.buffer[i] >> pos) & 1

			ch := off
			if bit == 1 {
				ch = on
			}

			start := x * s.scale
			for i := range s.scale {
				line[start+i] = ch
			}
		}

		for range s.scale {
			if _, err := w.Write([]byte(string(line))); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Screen) Clear() {
	for i := range s.buffer {
		s.buffer[i] = 0
	}
}
