package cpu

import (
	"github.com/valerio/go-jeebie/jeebie/addr"
	"github.com/valerio/go-jeebie/jeebie/bit"
	"github.com/valerio/go-jeebie/jeebie/memory"
)

// Flag is one of the 4 possible flags used in the flag register (high part of AF)
type Flag uint8

const (
	zeroFlag      Flag = 0x80
	subFlag            = 0x40
	halfCarryFlag      = 0x20
	carryFlag          = 0x10
)

const (
	baseInterruptAddress uint16 = 0x40
)

// Interrupt is an enum that represents one of the possible interrupts.
type Interrupt uint8

const (
	// VBlankInterrupt is fired when the GPU has completed a frame.
	VBlankInterrupt Interrupt = 1
	// LCDSTATInterrupt is fired based on one of the conditions in the LCDSTAT register.
	LCDSTATInterrupt = 1 << 1
	// TimerInterrupt is fired when the timer register (TIMA) overflows (i.e. goes from 0xFF to 0x00).
	TimerInterrupt = 1 << 2
	// SerialInterrupt is fired when a serial transfer has completed on the game link port.
	SerialInterrupt = 1 << 3
	// JoypadInterrupt is fired when any of the keypad inputs goes from high to low.
	JoypadInterrupt = 1 << 4
)

// CPU is the main struct holding Z80 state
type CPU struct {
	// registers
	a  uint8
	f  uint8
	b  uint8
	c  uint8
	d  uint8
	e  uint8
	h  uint8
	l  uint8
	sp uint16
	pc uint16

	// metadata
	interruptsEnabled bool
	currentOpcode     uint16
	stopped           bool
	cycles            uint64

	memory *memory.MMU
}

// New returns an uninitialized CPU instance
func New(memory *memory.MMU) *CPU {
	return &CPU{
		memory: memory,
	}
}

// Tick emulates a single step during the main loop for the cpu.
// Returns the amount of cycles that execution has taken.
func (c *CPU) Tick() int {
	c.handleInterrupts()

	instruction := Decode(c)
	cycles := instruction(c)
	c.cycles += uint64(cycles)

	return cycles
}

// handleInterrupts checks for an interrupt and handles it if necessary.
func (c *CPU) handleInterrupts() {
	if c.interruptsEnabled == false {
		return
	}

	// retrieve the two masks
	enabledInterruptsMask := c.memory.Read(addr.IE)
	firedInterrupts := c.memory.Read(addr.IF)

	// if zero, no interrupts that are enabled were fired
	if (enabledInterruptsMask & firedInterrupts) == 0 {
		return
	}

	for i := uint8(0); i < 5; i++ {
		if bit.IsSet(i, firedInterrupts) {
			// interrupt handlers are offset by 8
			// 0x40 - 0x48 - 0x50 - 0x58 - 0x60
			address := uint16(i)*8 + baseInterruptAddress

			// mark as handled by clearing the bit at i
			c.memory.Write(addr.IF, bit.Clear(i, firedInterrupts))

			// move PC to interrupt handler address
			c.pushStack(c.pc)
			c.pc = address

			// add cycles equivalent to a JMP.
			c.cycles += 20

			// disable interrupts
			c.interruptsEnabled = false

			// return on the first served interrupt.
			// will serve the next when interrupts are enabled again.
			return
		}
	}
}

// peekImmediate returns the byte at the memory address pointed by the PC
// this value is known as immediate ('n' in mnemonics), some opcodes use it as a parameter
func (c CPU) peekImmediate() uint8 {
	n := c.memory.Read(c.pc)
	return n
}

// peekImmediateWord returns the two bytes at the memory address pointed by PC and PC+1
// this value is known as immediate ('nn' in mnemonics), some opcodes use it as a parameter
func (c CPU) peekImmediateWord() uint16 {
	low := c.memory.Read(c.pc)
	high := c.memory.Read(c.pc + 1)

	return bit.Combine(low, high)
}

// peekSignedImmediate returns signed byte value at the memory address pointed by PC
// this value is known as immediate ('*' in mnemonics), some opcodes use it as a parameter
func (c CPU) peekSignedImmediate() int8 {
	return int8(c.peekImmediate())
}

// readImmediate acts similarly as its peek counterpart, but increments the PC once after reading
func (c *CPU) readImmediate() uint8 {
	n := c.peekImmediate()
	c.pc++
	return n
}

// readImmediateWord acts similarly as its peek counterpart, but increments the PC twice after reading
func (c *CPU) readImmediateWord() uint16 {
	nn := c.peekImmediateWord()
	c.pc += 2
	return nn
}

// readSignedImmediate acts similarly as its peek counterpart, but increments the PC once after reading
func (c *CPU) readSignedImmediate() int8 {
	n := c.peekSignedImmediate()
	c.pc++
	return n
}

func (c *CPU) setFlag(flag Flag) {
	c.f &= uint8(flag)
}

func (c *CPU) resetFlag(flag Flag) {
	c.f &= uint8(flag ^ 0xFF)
}

func (c CPU) isSetFlag(flag Flag) bool {
	return c.f&uint8(flag) != 0
}

// flagToBit will return 1 if the passed flag is set, 0 otherwise
func (c CPU) flagToBit(flag Flag) uint8 {
	if c.isSetFlag(flag) {
		return 1
	}

	return 0
}

func (c *CPU) setFlagToCondition(flag Flag, condition bool) {
	if condition {
		c.setFlag(flag)
	} else {
		c.resetFlag(flag)
	}
}

func (c *CPU) setBC(value uint16) {
	c.b = bit.High(value)
	c.c = bit.Low(value)
}

func (c CPU) getBC() uint16 {
	return bit.Combine(c.b, c.c)
}

func (c *CPU) setDE(value uint16) {
	c.d = bit.High(value)
	c.e = bit.Low(value)
}

func (c CPU) getDE() uint16 {
	return bit.Combine(c.d, c.d)
}

func (c *CPU) setHL(value uint16) {
	c.h = bit.High(value)
	c.l = bit.Low(value)
}

func (c CPU) getHL() uint16 {
	return bit.Combine(c.h, c.l)
}

func (c *CPU) setAF(value uint16) {
	c.a = bit.High(value)
	c.f = bit.Low(value)
}

func (c CPU) getAF() uint16 {
	return bit.Combine(c.a, c.f)
}
