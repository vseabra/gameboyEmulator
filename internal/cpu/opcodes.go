package cpu

// opcode is a struct that holds the metadata and handler function for an opcode
type opcode struct {
	mnemonic string
	size     uint16
	handler  func(context context) (duration int, err error) // duration in clock cycles, divide by 4 to get duration in machine cycles
}

// context is a struct that holds the operands of an instruction and a pointer to the CPU
type context struct {
	n8  byte   // immediate 8-bit data
	e8  int8   // immediate signed 8-bit data
	n16 uint16 // immediate little-endian 16-bit data
	cpu *CPU
}

// mergeBytesToUint16 merges two 8-bit numbers into a 16-bit number
func mergeBytesToUint16(hi, lo byte) uint16 {
	return (uint16(hi) << 8) | uint16(lo)
}

// splitUint16IntoBytes splits a 16-bit number into two 8-bit numbers
func splitUint16IntoBytes(num uint16) (hi, lo byte) {
	hi = byte(num >> 8)
	lo = byte(num & 0x00FF)

	return hi, lo
}

// load8bit loads an 8-bit value into a register
func load8bit(register *byte, value byte) {
	*register = value
}

// load16Bit loads a 16-bit value into a register
func load16Bit(hiRegister *byte, loRegister *byte, value uint16) {
	hi, lo := splitUint16IntoBytes(value)
	*hiRegister, *loRegister = hi, lo
}

// checkBit checks if a bit is set in a register and sets the flags accordingly
func (c *CPU) checkBit(bitPosition byte, register byte) {
	mask := byte(1 << bitPosition)
	isZero := (register & mask) == 0

	c.setFlag(ZERO, isZero)
	c.setFlag(SUBSTRACT, false)
	c.setFlag(HALF_CARRY, true)
}

// incrementRegister increments a register and sets the flags accordingly
func (c *CPU) incrementRegister(register *byte) {
	registerBefore := *register
	*register += byte(1)

	c.setFlag(ZERO, *register == 0)
	c.setFlag(SUBSTRACT, false)
	c.setFlag(HALF_CARRY, registerBefore&0x0F == 0x0F)
}

// decrementRegister decremetns a register and sets the flags accordingly
func (c *CPU) decrementRegister(register *byte) {
	registerBefore := *register
	*register -= byte(1)

	c.setFlag(ZERO, *register == 0)
	c.setFlag(SUBSTRACT, true)
	c.setFlag(HALF_CARRY, (registerBefore&0x0F) == 0x00)
}

// push pushes a value into the stack
func (c *CPU) push(value uint16) error {
	c.sp -= 2
	hi, lo := splitUint16IntoBytes(value)
	err := c.memoryBus.WriteToAddress(c.sp, []byte{lo, hi})

	return err
}

// pop pops a value from the stack and returns it
func (c *CPU) pop() (hi byte, lo byte, err error) {
	bytes, err := c.memoryBus.ReadFromAddress(c.sp, 2)

	if err != nil {
		return 0, 0, err
	}

	c.sp += 2
	return bytes[1], bytes[0], err
}

// callSubroutine calls a subroutine at the given address
func (c *CPU) callSubroutine(address uint16) error {
	// Write current pc to stack
	err := c.push(c.PC)

	if err != nil {
		return err
	}

	// JMP to address
	c.PC = address

	return nil
}

// rotateLeft rotates a register to the left by one and set's the shiftedBit to the carry flag
// the old carry flag becomes the 8th bit of the register, implements the behaviour of RL N
func (c *CPU) rotateLeft(register *byte) {
	var carryFlag byte

	if c.getFlag(CARRY) {
		carryFlag = 1
	}

	// capture the bit that will be shifted
	shiftedBit := (*register & 0x80) >> 7

	// shift the register to the left and put the old carry flag at the right
	*register = (*register << 1) | carryFlag

	// set the carryFlag to the new shiftedBit
	c.setFlag(CARRY, shiftedBit == 1)

	// Set other flags
	c.setFlag(ZERO, *register == 0)
	c.setFlag(SUBSTRACT, false)
	c.setFlag(HALF_CARRY, false)

}

// func (c *CPU) xor(cpu *CPU, a, b byte) byte {
//     result := a ^ b
//     c.setFlag(ZERO, result == 0)
//     // TODO Set other flags
//     return result
// }

// opcodeLookup is a map of opcodes to their handler functions and metadata
var opcodeLookup = map[byte]opcode{
	0x01: {mnemonic: "LD BC n16", size: 3, handler: func(ctx context) (int, error) {
		load16Bit(&ctx.cpu.B, &ctx.cpu.C, ctx.n16)

		return 12, nil
	}},

	0x21: {mnemonic: "LD HL n16", size: 3, handler: func(ctx context) (int, error) {
		load16Bit(&ctx.cpu.H, &ctx.cpu.L, ctx.n16)

		return 12, nil
	}},

	0x31: {mnemonic: "LD SP n16", size: 3, handler: func(ctx context) (int, error) {
		ctx.cpu.sp = ctx.n16

		return 12, nil
	}},

	0x32: {mnemonic: "LD [HL-],A", size: 1, handler: func(ctx context) (int, error) {
		hl := mergeBytesToUint16(ctx.cpu.H, ctx.cpu.L)

		err := ctx.cpu.memoryBus.WriteToAddress(hl, []byte{ctx.cpu.A})
		if err != nil {
			return 0, err
		}

		hl--
		ctx.cpu.H, ctx.cpu.L = splitUint16IntoBytes(hl)

		return 8, nil
	}},

	0x0E: {mnemonic: "LD C n8", size: 2, handler: func(ctx context) (int, error) {
		load8bit(&ctx.cpu.C, ctx.n8)

		return 8, nil
	}},

	0x3E: {mnemonic: "LD A n8", size: 2, handler: func(ctx context) (int, error) {
		load8bit(&ctx.cpu.A, ctx.n8)

		return 8, nil
	}},

	0xAF: {mnemonic: "XOR A,A", size: 1, handler: func(ctx context) (int, error) {
		ctx.cpu.A = 0
		ctx.cpu.setFlag(ZERO, true)

		return 4, nil
	}},

	0x20: {mnemonic: "JR NZ e8", size: 2, handler: func(ctx context) (int, error) {

		if !ctx.cpu.getFlag(ZERO) {
			offset := ctx.e8
			newPc := int16(ctx.cpu.PC) + int16(offset)

			ctx.cpu.PC = uint16(newPc)
			return 12, nil
		}

		return 8, nil
	}},

	0xE2: {mnemonic: "LD [C] A", size: 1, handler: func(ctx context) (int, error) {
		offset := uint16(0xFF00)
		address := offset + uint16(ctx.cpu.C)

		err := ctx.cpu.memoryBus.WriteToAddress(address, []byte{ctx.cpu.A})

		if err != nil {
			return 0, err
		}

		return 8, nil
	}},

	0x0C: {mnemonic: "INC C", size: 1, handler: func(ctx context) (int, error) {
		ctx.cpu.incrementRegister(&ctx.cpu.C)

		return 4, nil
	}},

	0x77: {mnemonic: "LD [HL] A", size: 1, handler: func(ctx context) (int, error) {
		hl := mergeBytesToUint16(ctx.cpu.H, ctx.cpu.L)
		err := ctx.cpu.memoryBus.WriteToAddress(hl, []byte{ctx.cpu.A})

		if err != nil {
			return 0, err
		}

		return 8, nil
	}},

	// TODO: test me
	0xE0: {mnemonic: "LDH (N), A", size: 2, handler: func(ctx context) (int, error) {
		offset := uint16(0xFF00)
		address := offset + uint16(ctx.n8)

		err := ctx.cpu.memoryBus.WriteToAddress(address, []byte{ctx.cpu.A})

		if err != nil {
			return 0, err
		}

		return 12, nil
	}},

	0x11: {mnemonic: "LD DE n16", size: 3, handler: func(ctx context) (int, error) {
		load16Bit(&ctx.cpu.D, &ctx.cpu.E, ctx.n16)

		return 12, nil
	}},

	0x1A: {mnemonic: "LD A [DE]", size: 1, handler: func(ctx context) (int, error) {
		DE := mergeBytesToUint16(ctx.cpu.D, ctx.cpu.E)
		bytes, err := ctx.cpu.memoryBus.ReadFromAddress(DE, 1)

		if err != nil {
			return 0, err
		}

		value := bytes[0]

		load8bit(&ctx.cpu.A, value)

		return 8, nil
	}},

	// TODO: test me
	0xCD: {mnemonic: "CALL n16", size: 3, handler: func(ctx context) (int, error) {
		return 24, ctx.cpu.callSubroutine(ctx.n16)
	}},

	0x4F: {mnemonic: "LD C A", size: 1, handler: func(ctx context) (int, error) {
		load8bit(&ctx.cpu.C, ctx.cpu.A)

		return 4, nil
	}},

	0x06: {mnemonic: "LD B n8", size: 2, handler: func(ctx context) (int, error) {
		load8bit(&ctx.cpu.B, ctx.n8)

		return 8, nil
	}},

	0xC5: {mnemonic: "PUSH BC", size: 1, handler: func(ctx context) (int, error) {
		err := ctx.cpu.push(mergeBytesToUint16(ctx.cpu.B, ctx.cpu.C))

		if err != nil {
			return 0, err
		}

		return 16, nil
	}},

	0x17: {mnemonic: "RLA", size: 1, handler: func(ctx context) (int, error) {
		ctx.cpu.rotateLeft(&ctx.cpu.A)

		ctx.cpu.setFlag(ZERO, false)

		return 4, nil
	}},

	0xC1: {mnemonic: "POP BC", size: 1, handler: func(ctx context) (int, error) {
		hi, lo, err := ctx.cpu.pop()

		if err != nil {
			return 0, err
		}

		ctx.cpu.B, ctx.cpu.C = hi, lo
		return 12, nil
	}},

	0x05: {mnemonic: "DEC B", size: 1, handler: func(ctx context) (int, error) {
		ctx.cpu.decrementRegister(&ctx.cpu.B)
		return 4, nil
	}},
}

// cbPrefixedOpcodeLookup is a map of opcodes to their handler functions and metadata
var cbPrefixedOpcodeLookup = map[byte]opcode{
	// TODO - tests
	0x7C: {mnemonic: "BIT 7 H", size: 2, handler: func(ctx context) (int, error) {
		ctx.cpu.checkBit(7, ctx.cpu.H)

		return 8, nil
	}},

	0x11: {mnemonic: "RL C", size: 2, handler: func(ctx context) (int, error) {
		ctx.cpu.rotateLeft(&ctx.cpu.C)
		return 8, nil
	}},
}

// RLCA (Rotate Left Circular A):
//     This is a rotate left instruction.
//     It rotates all the bits in the accumulator (A) to the left.
//     The old bit 7 (the most significant bit) is copied into both the carry flag and bit 0 (the least significant bit).
//     This operation is circular, meaning the bit that gets shifted out on one end is put back in on the other end.
//     RLCA only affects the carry flag (C) in the F register. The zero (Z), negative (N), and half carry (H) flags are reset.

// RLA (Rotate Left through Carry A):
//     This is also a rotate left instruction, but it includes the carry flag.
//     It rotates the bits in the accumulator (A) to the left, and the old bit 7 is moved to the carry flag.
//     The old value of the carry flag is moved to bit 0 of the accumulator.
//     Unlike RLCA, RLA is not circular because it involves the carry flag.
//     Similar to RLCA, RLA only affects the carry flag (C). The zero (Z), negative (N), and half carry (H) flags are reset.
//     This means that RLN = RL N followed by resetting the zero flag

// RL A (Rotate Left A through Carry):
//     This is very similar to RLA but differs in how it affects the flags.
//     Like RLA, it rotates the accumulator (A) left through the carry flag.
//     The key difference is in how it sets the flags. RL A sets the zero flag (Z) if the result is zero, resets the negative (N) and half carry (H) flags, and sets or clears the carry flag (C) depending on the rotation.
