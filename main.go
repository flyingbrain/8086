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
	{0b11111110, 0b11000110, "mov", immMovRegMemDecode},
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
	{mask: 0b11110000, value: 0b01110000, name: "none", decode: jumpDecode},
	{mask: 0b11111100, value: 0b11100000, name: "none", decode: loopDecode},
}

var jccTable = [...]string{
	"jo",  // 0000
	"jno", // 0001
	"jb",  // 0010
	"jnb", // 0011
	"je",  // 0100
	"jne", // 0101
	"jbe", // 0110
	"ja",  // 0111
	"js",  // 1000
	"jns", // 1001
	"jp",  // 1010
	"jnp", // 1011
	"jl",  // 1100
	"jge", // 1101
	"jle", // 1110
	"jg",  // 1111
}

var loopTable = map[byte]string{
	0b11100000: "loopne",
	0b11100001: "loope",
	0b11100010: "loop",
	0b11100011: "jcxz",
}

func getNumber(data []byte, signed bool) int {
	var u uint16

	if len(data) == 1 {
		u = uint16(data[0])
		if signed {
			return int(int8(u)) // correct sign for 1 byte
		}
		return int(u)
	}

	if len(data) == 2 {
		u = uint16(data[1])<<8 | uint16(data[0]) // little endian

		if signed {
			return int(int16(u)) // correct sign for 2 bytes
		}
		return int(u)
	}

	panic("unsupported size")
}

func jumpDecode(b byte, file io.ReadCloser) (string, error) {
	cond := b & 0b1111
	name := jccTable[cond]

	buf := make([]byte, 1)

	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	disp := int8(buf[0])

	return fmt.Sprintf("%s %d\n", name, disp), nil
}

func loopDecode(b byte, file io.ReadCloser) (string, error) {
	name := loopTable[b]

	buf := make([]byte, 1)

	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	disp := int8(buf[0])

	return fmt.Sprintf("%s %d\n", name, disp), nil
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

func immMovRegMemDecode(b byte, file io.ReadCloser) (string, error) {
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

	value := getNumber(buf, true)
	command := ""

	if w == 0b1 {
		command = fmt.Sprintf("%s, word %d\n", rmText, value)
	} else {
		command = fmt.Sprintf("%s, byte %d\n", rmText, value)
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

	command := commands[uint8(opt)]

	rmText := getRM(rm, w, mod, file)

	if s|w == 0b11 {
		buf = make([]byte, 2)
	}

	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf, s == 0b1)

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
	size := 1
	if w == 0b1 {
		size = 2
	}

	buf := make([]byte, size)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	reg := "al"
	value := getNumber(buf[:1], false)
	if w == 0b1 {
		value = getNumber(buf, false)
		reg = "ax"
	}

	command := ""

	if p == 0b1 {
		command = fmt.Sprintf("[%d], %s\n", value, reg)
	} else {
		command = fmt.Sprintf("%s,[%d]\n", reg, value)
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
