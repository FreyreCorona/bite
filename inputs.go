package main

import (
	"io"
	"math/bits"
)

type Keyboard struct {
	reader io.Reader
	keys   []uint16
	size   int
}

func NewKeyboard(size int, r io.Reader) *Keyboard {
	if size < 1 {
		size = 1
	}

	words := (size + 16 - 1) / 16
	return &Keyboard{
		keys:   make([]uint16, words),
		size:   size,
		reader: r,
	}
}

func (k *Keyboard) Run() error {
	if k.reader == nil {
		return nil
	}

	buf := make([]byte, 2) // [key, state]

	for {
		_, err := io.ReadFull(k.reader, buf)
		if err != nil {
			return err
		}

		key := int(buf[0])
		down := buf[1] == 1

		if down {
			k.Press(key)
		} else {
			k.Release(key)
		}
	}
}

func (k *Keyboard) Press(key int) {
	if key < 0 || key >= k.size {
		return
	}

	word := key / 16
	bit := uint16(key % 16)

	k.keys[word] |= uint16(1) << bit
}

func (k *Keyboard) Release(key int) {
	if key < 0 || key >= k.size {
		return
	}

	word := key / 16
	bit := uint16(key % 16)

	k.keys[word] &^= uint16(1) << bit
}

func (k *Keyboard) IsPressed(key int) bool {
	if key < 0 || key >= k.size {
		return false
	}

	word := key / 16
	bit := uint16(key % 16)

	return (k.keys[word] & (uint16(1) << bit)) != 0
}

func (k *Keyboard) AnyPressed() (int, bool) {
	for i, word := range k.keys {
		if word == 0 {
			continue
		}

		b := bits.TrailingZeros16(word)
		key := i*16 + b
		if key < k.size {
			return key, true
		}
		return 0, false
	}
	return 0, false
}
