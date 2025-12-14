package main

import (
	"flag"
	"io"
	"log"
	"os"
	"time"
)

var (
	ROM string

	Out    string
	ScrOut io.Writer = os.Stdout
	ScrW   int
	ScrH   int
	Scale  int

	In      string
	InputIn io.Reader = os.Stdin
	Keys    int

	Oa         string
	AudioOut   io.Writer
	SampleRate int
)

func parseFlags() {
	flag.StringVar(&ROM, "rom", "", "path to ROM file")

	flag.StringVar(&Out, "o", "", "screen output")
	flag.IntVar(&ScrW, "width", 64, "screen width")
	flag.IntVar(&ScrH, "height", 32, "screen height")
	flag.IntVar(&Scale, "scale", 1, "screen scale factor")

	flag.StringVar(&In, "i", "", "input device")
	flag.IntVar(&Keys, "keys", 16, "keyboard size")

	flag.StringVar(&Oa, "a", "", "audio output")
	flag.IntVar(&SampleRate, "samplerate", 44100, "audio sample rate")

	flag.Parse()

	if ROM == "" {
		log.Fatal("no ROM specified")
	}
}

func main() {
	parseFlags()

	if Out != "" {
		f, err := os.Create(Out)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		ScrOut = f
	}

	if Oa != "" {
		f, err := os.Create(Oa)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		AudioOut = f
	}

	if In != "" {
		f, err := os.Open(In)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		InputIn = f
	}

	screen := NewScreen(ScrW, ScrH, Scale, ScrOut)
	keyboard := NewKeyboard(Keys, InputIn)

	var audio *Audio
	if AudioOut != nil {
		audio = NewAudio(SampleRate, 1, 16, AudioOut)
	}

	cpu := NewCPU(screen, keyboard, audio)

	data, err := os.ReadFile(ROM)
	if err != nil {
		log.Fatal(err)
	}
	cpu.LoadROM(data)

	os.Exit(RunTTY(cpu, screen, keyboard, audio))
}

func RunTTY(cpu *CPU, screen *Screen, keyboard *Keyboard, audio *Audio) int {
	ScrOut.Write([]byte("\x1b[?25l"))

	go func() {
		if err := keyboard.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		cpu.Step()

		if ScrOut == os.Stdout {
			ScrOut.Write([]byte("\033[H"))
		}
		if err := screen.Render(); err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Millisecond * 16)
	}

	ScrOut.Write([]byte("\x1b[?25h"))
	return 0
}
