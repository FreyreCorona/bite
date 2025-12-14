package main

import (
	"fmt"
	"math/rand"
)

var chip8Font = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type CPU struct {
	Memory [4096]uint8
	V      [16]uint8
	I      uint16
	PC     uint16
	Stack  [16]uint16
	SP     uint8
	DT     uint8
	ST     uint8

	Display  *Screen
	Keyboard *Keyboard
	Audio    *Audio
}

func NewCPU(d *Screen, k *Keyboard, a *Audio) *CPU {
	c := &CPU{
		PC:       0x200,
		Display:  d,
		Keyboard: k,
		Audio:    a,
	}
	copy(c.Memory[0x050:], chip8Font[:])

	return c
}

func (c *CPU) LoadROM(data []byte) error {
	if len(data)+0x200 > len(c.Memory) {
		return fmt.Errorf("ROM too large: %d bytes", len(data))
	}

	copy(c.Memory[0x200:], data)
	c.PC = 0x200

	return nil
}

func (c *CPU) Step() error {
	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])

	c.PC += 2

	return c.execute(uint16(opcode))
}

func (c *CPU) TickTimers() {
	if c.DT > 0 {
		c.DT--
	}

	if c.ST > 0 {
		c.ST--
	}

	if c.Audio == nil {
		return
	}

	if c.ST > 0 {
		c.Audio.On()
	} else {
		c.Audio.Off()
	}
}

func (c *CPU) execute(op uint16) error {
	nibble := op & 0xF000

	switch nibble {
	case 0x0000:
		switch op {
		case 0x00E0:
			return c.cls()
		case 0x00EE:
			return c.ret()
		default:
			return c.sys(op)
		}
	case 0x1000:
		return c.jp(op & 0x0FFF)
	case 0x2000:
		return c.call(op & 0x0FFF)
	case 0x3000:
		return c.seByte(op)
	case 0x4000:
		return c.sneByte(op)
	case 0x5000:
		return c.seReg(op)
	case 0x6000:
		return c.ldByte(op)
	case 0x7000:
		return c.addByte(op)
	case 0x8000:
		switch op & 0x0000F {
		case 0x0:
			return c.ldReg(op)
		case 0x1:
			return c.or(op)
		case 0x2:
			return c.and(op)
		case 0x3:
			return c.xor(op)
		case 0x4:
			return c.addReg(op)
		case 0x5:
			return c.sub(op)
		case 0x6:
			return c.shr(op)
		case 0x7:
			return c.subn(op)
		case 0xE:
			return c.shl(op)
		}
	case 0x9000:
		return c.sneReg(op)
	case 0xA000:
		return c.ldI(op)
	case 0xB000:
		return c.jpV0(op)
	case 0xC000:
		return c.rnd(op)
	case 0xD000:
		return c.drw(op)
	case 0xE000:
		switch op & 0x00FF {
		case 0x9E:
			return c.skp(op)
		case 0xA1:
			return c.sknp(op)

		}
	case 0xF000:
		switch op & 0x00FF {
		case 0x07:
			return c.ldVxDT(op)
		case 0x0A:
			return c.ldVxKey(op)
		case 0x15:
			return c.ldDT(op)
		case 0x18:
			return c.ldST(op)
		case 0x1E:
			return c.addI(op)
		case 0x29:
			return c.ldSprite(op)
		case 0x33:
			return c.bcd(op)
		case 0x55:
			return c.store(op)
		case 0x65:
			return c.load(op)
		}
	}
	return nil
}

// 00E0: CLS
func (c *CPU) cls() error {
	c.Display.Clear()
	return nil
}

// 00EE: RET
func (c *CPU) ret() error {
	c.SP--
	c.PC = c.Stack[c.SP]
	return nil
}

// 0nnn: SYS addr (ignorado por la mayoría de intérpretes)
func (c *CPU) sys(_ uint16) error {
	return nil
}

// 1nnn: JP addr
func (c *CPU) jp(addr uint16) error {
	c.PC = addr
	return nil
}

// 2nnn: CALL addr
func (c *CPU) call(addr uint16) error {
	c.Stack[c.SP] = c.PC
	c.SP++
	c.PC = addr
	return nil
}

// 3xkk: SE Vx, byte
func (c *CPU) seByte(op uint16) error {
	x := (op >> 8) & 0xF
	kk := byte(op)
	if c.V[x] == kk {
		c.PC += 2
	}
	return nil
}

// 4xkk: SNE Vx, byte
func (c *CPU) sneByte(op uint16) error {
	x := (op >> 8) & 0xF
	kk := byte(op)
	if c.V[x] != kk {
		c.PC += 2
	}
	return nil
}

// 5xy0: SE Vx, Vy
func (c *CPU) seReg(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	if c.V[x] == c.V[y] {
		c.PC += 2
	}
	return nil
}

// 6xkk: LD Vx, byte
func (c *CPU) ldByte(op uint16) error {
	x := (op >> 8) & 0xF
	c.V[x] = byte(op)
	return nil
}

// 7xkk: ADD Vx, byte
func (c *CPU) addByte(op uint16) error {
	x := (op >> 8) & 0xF
	kk := byte(op)
	c.V[x] += kk
	return nil
}

// 8xy0: LD Vx, Vy
func (c *CPU) ldReg(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	c.V[x] = c.V[y]
	return nil
}

// 8xy1: OR Vx, Vy
func (c *CPU) or(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	c.V[x] |= c.V[y]
	return nil
}

// 8xy2: AND Vx, Vy
func (c *CPU) and(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	c.V[x] &= c.V[y]
	return nil
}

// 8xy3: XOR Vx, Vy
func (c *CPU) xor(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	c.V[x] ^= c.V[y]
	return nil
}

// 8xy4: ADD Vx, Vy (con carry)
func (c *CPU) addReg(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	sum := uint16(c.V[x]) + uint16(c.V[y])
	c.V[0xF] = 0
	if sum > 0xFF {
		c.V[0xF] = 1
	}
	c.V[x] = byte(sum)
	return nil
}

// 8xy5: SUB Vx, Vy
func (c *CPU) sub(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	if c.V[x] > c.V[y] {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[x] -= c.V[y]
	return nil
}

// 8xy6: SHR Vx
func (c *CPU) shr(op uint16) error {
	x := (op >> 8) & 0xF
	c.V[0xF] = c.V[x] & 1
	c.V[x] >>= 1
	return nil
}

// 8xy7: SUBN Vx, Vy
func (c *CPU) subn(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	if c.V[y] > c.V[x] {
		c.V[0xF] = 1
	} else {
		c.V[0xF] = 0
	}
	c.V[x] = c.V[y] - c.V[x]
	return nil
}

// 8xyE: SHL Vx
func (c *CPU) shl(op uint16) error {
	x := (op >> 8) & 0xF
	c.V[0xF] = (c.V[x] & 0x80) >> 7
	c.V[x] <<= 1
	return nil
}

// 9xy0: SNE Vx, Vy
func (c *CPU) sneReg(op uint16) error {
	x := (op >> 8) & 0xF
	y := (op >> 4) & 0xF
	if c.V[x] != c.V[y] {
		c.PC += 2
	}
	return nil
}

// Annn: LD I, addr
func (c *CPU) ldI(op uint16) error {
	c.I = op & 0x0FFF
	return nil
}

// Bnnn: JP V0, addr
func (c *CPU) jpV0(op uint16) error {
	c.PC = (op & 0x0FFF) + uint16(c.V[0])
	return nil
}

// Cxkk: RND Vx, byte
func (c *CPU) rnd(op uint16) error {
	x := (op >> 8) & 0xF
	kk := byte(op)
	r := byte(rand.Intn(256))
	c.V[x] = r & kk
	return nil
}

// Dxyn: DRW
func (c *CPU) drw(op uint16) error {
	vx := int(c.V[(op>>8)&0xF])
	vy := int(c.V[(op>>4)&0xF])
	n := int(op & 0xF)

	c.V[0xF] = 0

	for row := range n {
		sprByte := c.Memory[c.I+uint16(row)]
		y := (vy + row) % c.Display.Height

		for col := range 8 {
			if (sprByte & (0x80 >> col)) == 0 {
				continue
			}

			x := (vx + col) % c.Display.Width

			byteVal := c.Display.Get(x, y)
			pixel := (byteVal >> (7 - (x % 8))) & 1

			if pixel == 1 {
				c.V[0xF] = 1
			}

			c.Display.Set(x, y, pixel == 0)
		}
	}

	return nil
}

// Ex9E: SKP Vx
func (c *CPU) skp(op uint16) error {
	x := (op >> 8) & 0xF
	key := int(c.V[x])

	if c.Keyboard.IsPressed(key) {
		c.PC += 2
	}
	return nil
}

// ExA1: SKNP Vx
func (c *CPU) sknp(op uint16) error {
	x := (op >> 8) & 0xF
	key := int(c.V[x])

	if !c.Keyboard.IsPressed(key) {
		c.PC += 2
	}
	return nil
}

// Fx07: LD Vx, DT
func (c *CPU) ldVxDT(op uint16) error {
	x := (op >> 8) & 0xF
	c.V[x] = c.DT
	return nil
}

// Fx0A: LD Vx, K (espera tecla)
func (c *CPU) ldVxKey(op uint16) error {
	x := (op >> 8) & 0xF
	if key, pressed := c.Keyboard.AnyPressed(); pressed {
		c.V[x] = uint8(key)
		return nil
	}

	c.PC -= 2
	return nil
}

// Fx15: LD DT, Vx
func (c *CPU) ldDT(op uint16) error {
	x := (op >> 8) & 0xF
	c.DT = c.V[x]
	return nil
}

// Fx18: LD ST, Vx
func (c *CPU) ldST(op uint16) error {
	x := (op >> 8) & 0xF
	c.ST = c.V[x]
	return nil
}

// Fx1E: ADD I, Vx
func (c *CPU) addI(op uint16) error {
	x := (op >> 8) & 0xF
	c.I += uint16(c.V[x])
	return nil
}

// Fx29: LD F, Vx (sprite hexadecimal)
func (c *CPU) ldSprite(op uint16) error {
	x := (op >> 8) & 0xF
	c.I = uint16(c.V[x]) * 5
	return nil
}

// Fx33: BCD
func (c *CPU) bcd(op uint16) error {
	x := (op >> 8) & 0xF
	val := c.V[x]
	c.Memory[c.I] = val / 100
	c.Memory[c.I+1] = (val / 10) % 10
	c.Memory[c.I+2] = val % 10
	return nil
}

// Fx55: STORE V0..Vx en memoria
func (c *CPU) store(op uint16) error {
	x := (op >> 8) & 0xF
	for i := uint16(0); i <= x; i++ {
		c.Memory[c.I+i] = c.V[i]
	}
	return nil
}

// Fx65: LOAD V0..Vx desde memoria
func (c *CPU) load(op uint16) error {
	x := (op >> 8) & 0xF
	for i := uint16(0); i <= x; i++ {
		c.V[i] = c.Memory[c.I+i]
	}
	return nil
}

func (c *CPU) LoadProgram(data []byte) {
	copy(c.Memory[0x200:], data)
}
