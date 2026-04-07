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

	name += ".asm"

	asmfile, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}

	defer asmfile.Close()

	asmfile.WriteString("bits 16\n\n")

	buf := make([]byte, 2)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		if n != 2 {
			log.Fatal("wrong file format")
		}

		fmt.Printf("First byte %08b, second byte %08b \n", buf[0], buf[1])
		var command string

		opcode := fmt.Sprintf("%06b", buf[0]>>2)
		d := fmt.Sprintf("%01b", buf[0]<<6>>7)
		w := fmt.Sprintf("%01b", buf[0]<<7>>7)
		mod := fmt.Sprintf("%02b", buf[1]>>6)
		reg := fmt.Sprintf("%03b", buf[1]<<2>>5)
		r := fmt.Sprintf("%03b", buf[1]<<5>>5)

		fmt.Println("d: " + d)
		fmt.Println("w: " + w)
		fmt.Println("mod: " + mod)
		fmt.Println("reg: " + reg)
		fmt.Println("r/m: " + r)

		if opcode == "100010" {
			command += "mov"
		}

		command += " " + getRegister(r, w)
		command += ", " + getRegister(reg, w)
		command += "\n"

		asmfile.WriteString(command)
		fmt.Println(command)
	}
}

func getRegister(register string, w string) string {
	switch register {
	case "000":
		if w == "0" {
			return "al"
		}
		return "ax"
	case "001":
		if w == "0" {
			return "cl"
		}
		return "cx"
	case "010":
		if w == "0" {
			return "dl"
		}
		return "dx"
	case "011":
		if w == "0" {
			return "bl"
		}
		return "bx"
	case "100":
		if w == "0" {
			return "ah"
		}
		return "sp"
	case "101":
		if w == "0" {
			return "ch"
		}
		return "bp"
	case "110":
		if w == "0" {
			return "dh"
		}
		return "si"
	case "111":
		if w == "0" {
			return "bh"
		}
		return "di"
	}

	return ""
}
