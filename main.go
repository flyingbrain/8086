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
	{0b11111100, 0b11000110, "mov", immRegMemDecode},
	{0b11110000, 0b10110000, "mov", immRegDecode},
	{0b11111100, 0b10100000, "mov", accumDecode},
	{0b11111100, 0b10100010, "mov", accumDecode},
	{0b11111100, 0b00000000, "add", regMemDecode},
	{0b11111100, 0b10000000, "none", immRegMemDecode},
	{0b11111100, 0b00000100, "add", accumDecode},
	{0b11111100, 0b00101000, "sub", regMemDecode},
	{0b11111100, 0b00101100, "sub", accumDecode},
	{0b11111100, 0b00111000, "cmp", regMemDecode},
	{0b11111100, 0b00111100, "cmp", accumDecode},
}

func getNumber(data []byte, signed bool) int {
	i := uint16(data[0])
	if len(data) == 2 {
		i = uint16(data[1])<<8 | uint16(data[0])
	}

	if signed {
		return int(i)
	}

	return int(int16(i))
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
	commands := make([]string, 8)
	commands[0] = "add "
	commands[5] = "sub "
	commands[7] = "cmp "

	w := b & 0b1
	s := b >> 1 & 0b1

	buf := make([]byte, 1)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	b2 := buf[0]

	mod := (b2 >> 6) & 0b11
	opt := b2 >> 3 & 0b111
	rm := b2 & 0b111

	rmText := getRM(rm, w, mod, file)

	if s|w == 0b11 {
		buf = make([]byte, 2)
	}

	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf, s == 0b1)
	command := commands[uint8(opt)]

	if w == 0b1 {
		command += fmt.Sprintf("%s, word %d\n", rmText, value)
	} else {
		command += fmt.Sprintf("%s, byte %d\n", rmText, value)
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

	value := getNumber(buf, false)

	return fmt.Sprintf("%s, %d\n", regText, value), nil
}

func accumDecode(b byte, file io.ReadCloser) (string, error) {
	w := b & 0b1
	p := b >> 1 & 0b1

	buf := make([]byte, 2)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf[:1], false)
	if w == 0b1 {
		value = getNumber(buf, false)
	}
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

		intValue = getNumber(buf, false)
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

outer:
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
					break outer
				}

				if err != nil {
					fmt.Println(err)
				}

				if instruction.name != "none" {
					fmt.Printf("%s %s", instruction.name, command)
				} else {
					fmt.Printf("%s", command)
				}

				break
			}
		}
	}
}
