package main

import (
	"fmt"
	"io"
	"log"
)

func getRegister(reg byte, w byte) string {
	regs8 := [...]string{"al", "cl", "dl", "bl", "ah", "ch", "dh", "bh"}
	regs16 := [...]string{"ax", "cx", "dx", "bx", "sp", "bp", "si", "di"}

	if w == 0b1 {
		return regs16[int(reg)]
	}

	return regs8[int(reg)]
}

// TODO refactor to work with data array istad of file
func getRM(rm byte, w byte, mod byte, file io.ReadCloser) string {
	if mod == 0b11 {
		return getRegister(rm, w)
	}

	regs := [...]string{
		"bx + si",
		"bx + di",
		"bp + si",
		"bp + di",
		"si",
		"di",
		"bp",
		"bx",
	}

	bufSize := 2
	reg := regs[int(rm)]

	if mod == 0b00 {
		bufSize = 0
		if rm == 0b110 {
			bufSize = 2
			reg = ""
		}
	}

	if mod == 0b01 {
		bufSize = 1
	}

	intValue := 0

	if bufSize != 0 {
		buf := make([]byte, bufSize)
		_, err := file.Read(buf)

		if err == io.EOF {
			log.Fatal("command is invalid")
		}

		intValue = getNumber(buf, true)
	}

	if intValue == 0 {
		return fmt.Sprintf("[%s]", reg)
	} else if reg == "" {
		return fmt.Sprintf("[%d]", intValue)
	}

	if intValue > 0 {
		return fmt.Sprintf("[%s + %d]", reg, intValue)
	} else {
		return fmt.Sprintf("[%s - %d]", reg, -intValue)
	}
}
