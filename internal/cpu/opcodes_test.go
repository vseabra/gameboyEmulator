package cpu

import (
	"testing"

	. "github.com/carvhal/gby/internal/testutils"
)

type mockMemController struct {
	ram []byte
}

func (c *mockMemController) ReadFromAddress(address uint16, ammount int) ([]byte, error) {
	return c.ram[address : address+uint16(ammount)], nil
}

func (c *mockMemController) WriteToAddress(address uint16, bytes []byte) error {
	copy(c.ram[address:], bytes)

	return nil
}

func mockMemory(state []byte) *mockMemController {
	return &mockMemController{ram: state}
}

func getMockCPU() *CPU {
	return &CPU{
		memoryBus: mockMemory(make([]byte, 0xFFFF)),
	}
}

func TestOpcodes(t *testing.T) {
	tests := []struct {
		name     string
		opcode   byte
		cycles   int
		prefixed bool
		setup    func(*context)
		expected func(*testing.T, context)
	}{
		{
			name:   "LD BC n16",
			opcode: 0x01,
			cycles: 12,
			setup:  func(c *context) { c.n16 = 0xABCD },
			expected: func(t *testing.T, c context) {
				Expect(t, c.cpu.B, "Hi byte").ToEqual(byte(0xAB))
				Expect(t, c.cpu.C, "Lo byte").ToEqual(byte(0xCD))
			},
		},
		{
			name:   "LD HL n16",
			opcode: 0x21,
			cycles: 12,
			setup:  func(c *context) { c.n16 = 0x5678 },
			expected: func(t *testing.T, c context) {
				HL := mergeBytesToUint16(c.cpu.H, c.cpu.L)
				Expect(t, HL, "HL").ToEqual(uint16(0x5678))
			},
		},
		{
			name:     "LD SP n16",
			opcode:   0x31,
			cycles:   12,
			setup:    func(c *context) { c.n16 = 0x5678 },
			expected: func(t *testing.T, c context) { Expect(t, c.cpu.sp, "SP").ToEqual(uint16(0x5678)) },
		},
		{
			name:   "LD [HL-],A",
			opcode: 0x32,
			cycles: 8,
			setup: func(c *context) {
				c.cpu = &CPU{memoryBus: mockMemory(make([]byte, 0xDFFF)), H: 0xBB, L: 0xCC, A: 0x10}
			},
			expected: func(t *testing.T, c context) {
				HL := mergeBytesToUint16(c.cpu.H, c.cpu.L)

				// HL should be decremented
				Expect(t, HL, "HL").ToEqual(uint16(0xBBCB))

				// [HL] should be equal to A
				HLPointer := HL + 1
				HLValue, _ := c.cpu.memoryBus.ReadFromAddress(HLPointer, 1)
				Expect(t, HLValue[0], "[HL]").ToEqual(c.cpu.A)

			},
		},
		{
			name:   "XOR A,A",
			opcode: 0xAF,
			cycles: 4,
			setup:  func(c *context) { c.cpu.A = 0xFF; c.cpu.F = 0b0111_0000 },
			expected: func(t *testing.T, c context) {
				Expect(t, c.cpu.A, "A").ToEqual(byte(0))

				zeroFlag := c.cpu.F & flagMap[ZERO]
				Expect(t, zeroFlag, "Zero Flag").ToEqual(flagMap[ZERO])
			},
		},
		{
			name:     "JR NZ, e8 - branched",
			opcode:   0x20,
			cycles:   12,
			setup:    func(c *context) { c.e8 = -22; c.cpu.F = ^flagMap[ZERO]; c.cpu.PC = 22 },
			expected: func(t *testing.T, c context) { Expect(t, c.cpu.PC, "PC").ToEqual(uint16(0)) },
		},
		{
			name:     "JR NZ, e8 - not branched",
			opcode:   0x20,
			cycles:   8,
			setup:    func(c *context) { c.e8 = -22; c.cpu.F = flagMap[ZERO]; c.cpu.PC = 22 },
			expected: func(t *testing.T, c context) { Expect(t, c.cpu.PC, "PC").ToEqual(uint16(22)) },
		},
		{
			name:   "LD [C], A",
			opcode: 0xE2,
			cycles: 8,
			setup:  func(c *context) { c.cpu.C = 0xAA; c.cpu.A = 0xFF },
			expected: func(t *testing.T, c context) {
				Expect(t, uint16(c.cpu.C), "C").ToEqual(uint16(0xAA))

				offset := uint16(0xFF00)
				bytes, _ := c.cpu.memoryBus.ReadFromAddress(uint16(c.cpu.C)+offset, 1)
				valueAtC := uint16(bytes[0])

				Expect(t, valueAtC, "[C]").ToEqual(uint16(0xFF))

			},
		},
		{
			name:   "INC C (overflow)",
			opcode: 0x0C,
			cycles: 4,
			setup:  func(c *context) { c.cpu.F = 0b1111_0000; c.cpu.C = 255 },
			expected: func(t *testing.T, c context) {
				Expect(t, c.cpu.C, "C").ToEqual(byte(0))
				Expect(t, c.cpu.F&flagMap[ZERO], "Zero flag").ToEqual(flagMap[ZERO])
				Expect(t, c.cpu.F&flagMap[SUBSTRACT], "Subtract flag").ToEqual(byte(0))
				Expect(t, c.cpu.F&flagMap[HALF_CARRY], "Half carry flag").ToEqual(flagMap[HALF_CARRY])
			},
		},
		{
			name:   "INC C",
			opcode: 0x0C,
			cycles: 4,
			setup:  func(c *context) { c.cpu.F = 0b1111_0000; c.cpu.C = 254 },
			expected: func(t *testing.T, c context) {
				Expect(t, c.cpu.C, "C").ToEqual(byte(255))
				Expect(t, c.cpu.F&flagMap[ZERO], "Zero flag").ToEqual(byte(0))
				Expect(t, c.cpu.F&flagMap[SUBSTRACT], "Subtract flag").ToEqual(byte(0))
				Expect(t, c.cpu.F&flagMap[HALF_CARRY], "Half carry flag").ToEqual(byte(0))
			},
		},
		{
			name:   "LD [HL], A",
			opcode: 0x77,
			cycles: 8,
			setup:  func(c *context) { c.cpu.H = 0xAA; c.cpu.L = 0xFF; c.cpu.A = 0x11 },
			expected: func(t *testing.T, c context) {
				hl := mergeBytesToUint16(c.cpu.H, c.cpu.L)
				Expect(t, hl, "HL").ToEqual(uint16(0xAAFF)) // HL is not changed

				valueAtHL, _ := c.cpu.memoryBus.ReadFromAddress(hl, 1)
				Expect(t, uint16(valueAtHL[0]), "[HL]").ToEqual(uint16(c.cpu.A))

			},
		},
		{
			name:     "RL C - Rotate Left with Carry",
			opcode:   0x11,
			prefixed: true,
			setup: func(ctx *context) {
				// Setup initial state: C = 0x85 (10000101 in binary), Carry Flag = 0
				ctx.cpu.C = 0x85
				ctx.cpu.setFlag(CARRY, false)
			},
			expected: func(t *testing.T, ctx context) {
				Expect(t, ctx.cpu.C, "C register").ToEqual(byte(0x0A))
				Expect(t, ctx.cpu.getFlag(CARRY), "Carry flag").ToEqual(true)
				Expect(t, ctx.cpu.getFlag(ZERO), "Zero flag").ToEqual(false)
				Expect(t, ctx.cpu.getFlag(SUBSTRACT), "Subtract flag").ToEqual(false)
				Expect(t, ctx.cpu.getFlag(HALF_CARRY), "Half-Carry flag").ToEqual(false)
			},
			cycles: 8,
		},
		{
			name:   "DEC B - Non-Zero Value",
			opcode: 0x05,
			cycles: 4,
			setup:  func(c *context) { c.cpu.B = 0x10 },
			expected: func(t *testing.T, c context) {
				Expect(t, c.cpu.B, "B register").ToEqual(byte(0x0F))
				Expect(t, c.cpu.getFlag(ZERO), "Zero flag").ToEqual(false)
			},
		},
		{
			name:   "DEC B - Half-Carry Flag",
			opcode: 0x05,
			cycles: 4,
			setup:  func(c *context) { c.cpu.B = 0x10 },
			expected: func(t *testing.T, c context) {
				Expect(t, c.cpu.getFlag(HALF_CARRY), "Half-Carry flag").ToEqual(true)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := context{cpu: getMockCPU()}
			test.setup(&context)

			var (
				cyclesTaken int
				err         error
			)

			if test.prefixed {
				cyclesTaken, err = cbPrefixedOpcodeLookup[test.opcode].handler(context)
			} else {
				cyclesTaken, err = opcodeLookup[test.opcode].handler(context)
			}

			Must(t, err, "Expected no error")
			Expect(t, cyclesTaken, "Cycle count").ToEqual(test.cycles)

			test.expected(t, context)
		})
	}

}
