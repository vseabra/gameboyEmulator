package cpu

import (
	"fmt"
)

func (c *CPU) fetch() (*instruction, error) {

	var lookup map[byte]opcode = opcodeLookup

	// fetch the opcode
	opcodeSlice, err := c.memoryBus.ReadFromAddress(c.PC, 1)

	if err != nil {
		return nil, fmt.Errorf("cpu error on fetch: %w", err)
	}

	// handle CB prefixed opcodes
	isPrefixed := opcodeSlice[0] == 0xCB

	if isPrefixed {
		lookup = cbPrefixedOpcodeLookup
		opcodeSlice, err = c.memoryBus.ReadFromAddress(c.PC+1, 1)

		if err != nil {
			return nil, fmt.Errorf("cpu error on fetch: %w", err)
		}
	}

	opcode, ok := lookup[opcodeSlice[0]]

	if !ok {
		return nil, fmt.Errorf("cpu error on fetch: illegal / unknown opcode 0x%X prefixed: %v", opcodeSlice[0], isPrefixed)
	}

	// build context for the instruction
	context := context{cpu: c}

	operands, err := c.memoryBus.ReadFromAddress(c.PC+1, int(opcode.size))
	if err != nil {
		return nil, fmt.Errorf("cpu error on fetch: %w", err)
	}

	switch opcode.size {
	case 1:
	case 2:
		context.n8 = operands[0]
		context.e8 = int8(operands[0])
	case 3:
		context.n16 = mergeBytesToUint16(operands[1], operands[0])
	}

	c.PC += opcode.size

	return &instruction{opcode: &opcode, context: context}, nil

}
