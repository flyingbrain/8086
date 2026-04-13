package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

type instr struct {
	mask   byte
	value  byte
	name   string
	decode func(b byte, file io.ReadCloser) (string, error)
}

var instructions = []instr{
	{0b11111100, 0b10001000, "mov", regMemDecode},
	{0b11110000, 0b10110000, "mov", immRegDecode},
	{0b11111100, 0b11000110, "mov", immRegMemDecode},
	{0b11111100, 0b10100000, "mov", accumDecode},
	{0b11111100, 0b10100010, "mov", accumDecode},
}

func getNumber(data []byte) uint16 {
	if len(data) == 2 {
		return uint16(data[1])<<8 | uint16(data[0])
	}

	return uint16(data[0])
}

func regMemDecode(b byte, file io.ReadCloser) (string, error) {
	d := (b >> 1) & 0b1
	w := b & 0b1

	buf := make([]byte, 1)

	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	b2 := buf[0]

	mod := (b2 >> 6) & 0b11
	reg := (b2 >> 3) & 0b111
	rm := b2 & 0b111

	rmText := getRM(rm, w, mod, file)
	regText := getRegister(reg, w)
	command := ""

	if d == 0b0 {
		command = fmt.Sprintf("%s, %s\n", rmText, regText)
	} else {
		command = fmt.Sprintf("%s, %s\n", regText, rmText)
	}

	return command, nil
}

func immRegMemDecode(b byte, file io.ReadCloser) (string, error) {
	w := b & 0b1

	buf := make([]byte, 1)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	b2 := buf[0]

	mod := (b2 >> 6) & 0b11
	rm := b2 & 0b111

	rmText := getRM(rm, w, mod, file)

	if w == 0b1 {
		buf = make([]byte, 2)
	}

	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf)
	command := ""

	if w == 0b1 {
		command = fmt.Sprintf("%s, word %d\n", rmText, value)
	} else {
		command = fmt.Sprintf("%s, byte %d\n", rmText, value)
	}

	return command, nil
}

func immRegDecode(b byte, file io.ReadCloser) (string, error) {
	w := b >> 3 & 0b1
	reg := b & 0b111

	regText := getRegister(reg, w)
	buf := make([]byte, 1)

	if w == 0b1 {
		buf = make([]byte, 2)
	}

	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf)

	return fmt.Sprintf("%s, %d\n", regText, value), nil
}

func accumDecode(b byte, file io.ReadCloser) (string, error) {
	p := b >> 1 & 0b1

	buf := make([]byte, 2)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf)
	command := ""

	if p == 0b1 {
		command = fmt.Sprintf("[%d], ax\n", value)
	} else {
		command = fmt.Sprintf("ax,[%d]\n", value)
	}

	return command, err
}

func getRegister(reg byte, w byte) string {
	regs8 := [...]string{"al", "cl", "dl", "bl", "ah", "ch", "dh", "bh"}
	regs16 := [...]string{"ax", "cx", "dx", "bx", "sp", "bp", "si", "di"}

	if w == 0b1 {
		return regs16[int(reg)]
	}

	return regs8[int(reg)]
}

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
		if rm == 0b110 {
			reg = ""
		} else {
			bufSize = 0
		}
	}

	if mod == 0b01 {
		bufSize = 1
	}

	intValue := int16(0)

	if bufSize != 0 {
		buf := make([]byte, bufSize)
		_, err := file.Read(buf)

		if err == io.EOF {
			log.Fatal("command is invalid")
		}

		intValue = int16(int8(buf[0]))

		if bufSize == 2 {
			intValue = int16(uint16(buf[1])<<8 | uint16(buf[0]))
		}
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

	for {
		buf := make([]byte, 1)
		_, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		b1 := buf[0]

		for _, instruction := range instructions {
			if b1&instruction.mask == instruction.value {
				command, err := instruction.decode(b1, file)

				if err == io.EOF {
					break
				}

				if err != nil {
					fmt.Println(err)
				}

				fmt.Printf("%s %s", instruction.name, command)
			}
		}
	}
}
