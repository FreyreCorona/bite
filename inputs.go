package main

import (
	"io"
	"math/bits"
	"sync"
	"time"
)

type Keyboard struct {
	reader io.Reader
	keys   []uint16
	size   int
	mu     sync.RWMutex
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

	keyMap := map[byte]int{
		'1': 0x1, '2': 0x2, '3': 0x3, '4': 0xC,
		'q': 0x4, 'w': 0x5, 'e': 0x6, 'r': 0xD,
		'a': 0x7, 's': 0x8, 'd': 0x9, 'f': 0xE,
		'z': 0xA, 'x': 0x0, 'c': 0xB, 'v': 0xF,
	}

	buf := make([]byte, 1)

	for {
		n, err := k.reader.Read(buf)
		if err != nil {
			return err
		}

		if n != 1 {
			continue
		}

		key, ok := keyMap[buf[0]]
		if !ok {
			continue
		}

		k.Press(int(key))
		go func(ke int) {
			time.Sleep(30 * time.Millisecond)
			k.Release(ke)
		}(key)
	}
}

func (k *Keyboard) Press(key int) {
	if key < 0 || key >= k.size {
		return
	}

	word := key / 16
	bit := uint16(key % 16)

	k.mu.Lock()
	k.keys[word] |= uint16(1) << bit
	k.mu.Unlock()
}

func (k *Keyboard) Release(key int) {
	if key < 0 || key >= k.size {
		return
	}

	word := key / 16
	bit := uint16(key % 16)

	k.mu.Lock()
	k.keys[word] &^= uint16(1) << bit
	k.mu.Unlock()
}

func (k *Keyboard) IsPressed(key int) bool {
	if key < 0 || key >= k.size {
		return false
	}

	word := key / 16
	bit := uint16(key % 16)

	k.mu.RLock()
	defer k.mu.RUnlock()

	return (k.keys[word] & (uint16(1) << bit)) != 0
}

func (k *Keyboard) AnyPressed() (int, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

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
