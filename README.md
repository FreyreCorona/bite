# bite

**bite** is a **dependency-free Chip-8 emulator** written in **pure Go**.

The goal of this project is to provide a clean, minimal, and understandable Chip-8 implementation, suitable for learning emulator development, experimenting with low-level systems, or serving as a base for further extensions.

---

## 🧠 What is Chip-8?

Chip-8 is a simple interpreted programming language created in the 1970s for 8-bit systems.  
Due to its small instruction set and straightforward architecture, it is commonly used as a first emulator project to understand CPU cycles, memory, graphics rendering, and input handling.

---

## 🛠️ Features

- Pure Go implementation (zero external dependencies)
- Complete Chip-8 CPU emulation
- Basic graphics rendering
- Keyboard input handling
- Modular structure (CPU, screen, input separation)
- Designed for readability and extensibility

---

## 🚀 Getting Started

### Requirements

- Go 1.25 or newer

### Build

```bash
git clone https://github.com/FreyreCorona/bite.git
cd bite
go build -o bite
```

### Run

```bash
./bite -rom path/to/rom.ch8
```

### License

This project is licensed under the MIT License.
See the LICENSE file for details.

### Author
Created by Freyre Corona — a lightweight Chip-8 emulator built to explore how classic systems work at the CPU and memory level.
