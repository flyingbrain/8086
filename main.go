package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Please set source the file")
	}
	name := args[1]

	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	fmt.Print("bits 16\n\n")

	buf := make([]byte, 2)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		if n != 2 {
			log.Fatal("wrong file format")
		}

		b1 := buf[0]
		b2 := buf[1]
		var command string

		opcode := b1 >> 2
		// d
		_ = (b1 >> 1) & 0b1
		w := b1 & 0b1

		// mod
		_ = (b2 >> 6) & 0b11
		reg := (b2 >> 3) & 0b111
		rm := b2 & 0b111

		if opcode == 0b100010 {
			command += "mov"
		}

		command += " " + getRegister(rm, w)
		command += ", " + getRegister(reg, w)
		command += "\n"

		fmt.Print(command)
	}
}

func getRegister(reg byte, w byte) string {
	regs8 := [...]string{"al", "cl", "dl", "bl", "ah", "ch", "dh", "bh"}
	regs16 := [...]string{"ax", "cx", "dx", "bx", "sp", "bp", "si", "di"}

	if w == 0b0 {
		return regs16[int(reg)]
	}

	return regs8[int(reg)]
}
