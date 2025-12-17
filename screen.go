package main

import (
	"io"
)

type Screen struct {
	Width    int
	Height   int
	rowBytes int
	buffer   []byte
	writer   io.Writer
}

func NewScreen(w, h int, writer io.Writer) *Screen {
	rowBytes := (w + 7) / 8
	buf := make([]byte, rowBytes*h)

	return &Screen{
		Width:    w,
		Height:   h,
		rowBytes: rowBytes,
		buffer:   buf,
		writer:   writer,
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

func (s *Screen) Get(x, y int) bool {
	if x < 0 || y < 0 || x >= s.Width || y >= s.Height {
		return false
	}

	i := y*s.rowBytes + (x / 8)
	pos := 7 - (x % 8)

	return (s.buffer[i] & (1 << pos)) != 0
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

func (s *Screen) Render(w, h int) error {
	maxW := min(s.Width, w)
	maxH := min(min(s.Height, h*2), s.Height)

	for y := 0; y < maxH; y += 2 {
		for x := range maxW {
			top := s.Get(x, y)
			bottom := y+1 < s.Height && s.Get(x, y+1)

			var ch rune
			switch {
			case top && bottom:
				ch = '█'
			case top:
				ch = '▀'
			case bottom:
				ch = '▄'
			default:
				ch = ' '
			}

			if _, err := s.writer.Write([]byte(string(ch))); err != nil {
				return err
			}

		}
		if _, err := s.writer.Write([]byte("\r\n")); err != nil {
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
