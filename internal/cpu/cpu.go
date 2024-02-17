package cpu

import (
	"fmt"

	"github.com/carvhal/gby/internal/common"
)

type register uint8

type flag uint8

const (
	A register = iota
	F
	B
	C
	D
	E
	H
	L
)

const (
	ZERO flag = iota
	SUBSTRACT
	HALF_CARRY
	CARRY
)

var flagMap = map[flag]byte{
	ZERO:       0b1000_0000,
	SUBSTRACT:  0b0100_0000,
	HALF_CARRY: 0b0010_0000,
	CARRY:      0b0001_0000,
}

// instruction embeds an opcode and the context(operands) of an instruction
type instruction struct {
	opcode *opcode
	context
}

// execute runs the instruction and returns the number of cycles it took and an error if any
func (i *instruction) execute() (cycle int, err error) {
	return i.opcode.handler(i.context)
}

type CPU struct {
	A         byte   // Accumulator
	F         byte   // Flag Register (stored in the hi nibble as ZNHC)
	B, C      byte   // Register pair BC
	D, E      byte   // Register pair DE
	H, L      byte   // Register pair HL
	sp        uint16 // Stack Pointer
	PC        uint16 // Program Counter
	memoryBus common.MemoryReadWriter
	callStack []*instruction
}

func NewCPU(memoryReadWriter common.MemoryReadWriter) *CPU {
	return &CPU{memoryBus: memoryReadWriter}
}

// Tick fetches the next instuction and executes it, returning the number of cycles it took and an error if any
func (c *CPU) Tick() (cycles int, err error) {
	instruction, err := c.fetch()
	fmt.Printf("0X%X\n", c.PC)

	if err != nil {
		return 0, err
	}

	c.callStack = append(c.callStack, instruction)

	return instruction.execute()
}

// PrintStack prints the opcodes and their contexts on stdout in the order they were called
func (c *CPU) PrintStack() {
	for i, instruction := range c.callStack {
		fmt.Printf("[%d] opcode: %s (n8: 0x%X, e8: %v, n16: 0x%X) \n", i, instruction.opcode.mnemonic, instruction.context.n8, instruction.context.e8, instruction.context.n16)
	}
}

// setFlag sets the state of a flag
func (c *CPU) setFlag(f flag, state bool) {
	if state {
		c.F = c.F | flagMap[f]
	} else {
		c.F = c.F & ^flagMap[f]
	}
}

// getFlag returns the state of a flag
func (c *CPU) getFlag(f flag) bool {
	return c.F&flagMap[f] == flagMap[f]
}

// func (c *CPU) flipFlag(f flag) {
// 	c.F = c.F ^ flagMap[f]
// }
