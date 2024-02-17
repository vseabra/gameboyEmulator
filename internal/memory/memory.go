package memory

import (
	"fmt"
)

/*
* Memory map
*
* start  | end    | description
*
* 0x0000 | 0x3FFF | cartridge ROM
* 0x4000 | 0x7FFF | cartridge(switchable bank) ROM
* 0x8000 | 0x9FFF | VRAM
* 0xA000 | 0xBFFF | cartrige RAM
* 0xC000 | 0xDFFF | work RAM
* 0xE000 | 0xFDFF | echo RAM
* 0xFE00 | 0xFE9F | object attribute memory
* 0xFEA0 | 0xFEFF | not usable
* 0xFE00 | 0xFF7F | I/O registers
* 0xFF80 | 0xFFFE | HRAM
* 0xFFFF | 0xFFFF | Interupt enable register
*
 */

// Controller is a struct that represents the memory controller/bus
// it implements the MemoryReadWriter interface
type Controller struct {
	cartridge []byte
	ram       []byte
	vram      []byte
	hram      []byte
}

func NewController(game []byte) *Controller {
	return &Controller{
		cartridge: game,
		vram:      make([]byte, 1024*8),
		ram:       make([]byte, 1024*8),
		hram:      make([]byte, 127),
	}
}

func toRAMSpace(address uint16) uint16 {
	return address - 0xC000
}

func toVRAMSpace(address uint16) uint16 {
	return address - 0x8000
}

func toHRAMSpace(address uint16) uint16 {
	return address - 0xFF80
}

func (c *Controller) ReadFromAddress(address uint16, ammount int) ([]byte, error) {
	switch {

	// cartridge
	case address <= 0x3FFF:
		return c.cartridge[address : address+uint16(ammount)], nil

	// VRAM
	case address >= 0x8000 && address <= 0x9FFF:
		mappedAddress := toVRAMSpace(address)
		return c.vram[mappedAddress : mappedAddress+uint16(ammount)], nil

	// work RAM
	case address >= 0xC000 && address <= 0xDFFF:
		mappedAddress := toRAMSpace(address)
		return c.ram[mappedAddress : mappedAddress+uint16(ammount)], nil

	// HRAM
	case address >= 0xFF80 && address <= 0xFFFE:
		mappedAddress := toHRAMSpace(address)
		return c.hram[mappedAddress : mappedAddress+uint16(ammount)], nil

	}

	return nil, fmt.Errorf("Illegal read at 0x%X", address)
}

func (c *Controller) WriteToAddress(address uint16, bytes []byte) error {

	switch {

	// cartridge
	case address <= 0x3FFF:
		return fmt.Errorf("write to ROM (index 0x%X)", address)

	// VRAM
	case address >= 0x8000 && address <= 0x9FFF:
		mappedAddress := toVRAMSpace(address)
		copy(c.vram[mappedAddress:], bytes)
		return nil

	// work RAM
	case address >= 0xC000 && address <= 0xDFFF:
		mappedAddress := toRAMSpace(address)
		copy(c.ram[mappedAddress:], bytes)
		return nil

	// HRAM
	case address >= 0xFF80 && address <= 0xFFFE:
		mappedAddress := toHRAMSpace(address)
		copy(c.hram[mappedAddress:], bytes)
		return nil

	// IO
	case address >= 0xFE00 && address <= 0xFF7F:
		// TODO
		return nil
	}

	return fmt.Errorf("Illegal write at 0x%X", address)
}
