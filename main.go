package main

import (
	"flag"
	"io"
	"log"
	"os"
	"time"
)

var (
	ROM       string
	Out       string
	ScreenOut io.Writer = os.Stdout
	In        string
	InputIn   io.Reader = os.Stdin
	Oa        string
	AudioOut  io.Writer

	ScrW  int
	ScrH  int
	Scale int
	Keys  int

	SampleRate int
)

func parseFlags() {
	flag.StringVar(&ROM, "rom", "", "path to ROM file")

	flag.StringVar(&Out, "o", "", "screen output")
	flag.IntVar(&ScrW, "w", 64, "screen width")
	flag.IntVar(&ScrH, "h", 32, "screen height")
	flag.IntVar(&Scale, "scale", 1, "screen scale factor")

	flag.StringVar(&In, "i", "", "input device")
	flag.IntVar(&Keys, "keys", 16, "keyboard size")

	flag.StringVar(&Oa, "oa", "", "audio output")
	flag.IntVar(&SampleRate, "sr", 44100, "audio sample rate")

	flag.Parse()

	if ROM == "" {
		log.Fatal("no ROM specified")
	}

	if Out != "" {
		f, err := os.Create(Out)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		ScreenOut = f
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
		f, err := os.Create(In)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		InputIn = f
	}
}

func main() {
	parseFlags()

	screen := NewScreen(ScrW, ScrH, Scale, ScreenOut)
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
	ScreenOut.Write([]byte("\x1b[?25l"))

	go func() {
		if err := keyboard.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		cpu.Step()

		if ScreenOut == os.Stdout {
			ScreenOut.Write([]byte("\033[H"))
		}
		if err := screen.Render(); err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Millisecond * 16)
	}

	ScreenOut.Write([]byte("\x1b[?25h"))
	return 0
}
