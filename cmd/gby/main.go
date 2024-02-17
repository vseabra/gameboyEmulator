package main

import (
	"fmt"
	"os"

	"github.com/carvhal/gby/internal/cpu"
	"github.com/carvhal/gby/internal/memory"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: gby <rom>")
		os.Exit(1)
	}

	rom, err := os.ReadFile(os.Args[1])

	if err != nil {
		panic(err)
	}

	bus := memory.NewController(rom)
	cpu := cpu.NewCPU(bus)

	for {
		_, err := cpu.Tick()
		if err != nil {
			cpu.PrintStack()
			fmt.Printf("\nFATAL ERROR: %v at PC: 0x%X\nprinted call stack and exited... \n\n\n", err, cpu.PC)
			os.Exit(1)
		}
	}

}
