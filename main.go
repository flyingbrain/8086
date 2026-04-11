package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

	for {
		buf := make([]byte, 1)
		_, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		b1 := buf[0]

		command := "mov"

		if b1>>2 == 0b100010 {
			d := (b1 >> 1) & 0b1
			w := b1 & 0b1

			_, err := file.Read(buf)
			if err == io.EOF {
				break
			}

			b2 := buf[0]

			mod := (b2 >> 6) & 0b11
			reg := (b2 >> 3) & 0b111
			rm := b2 & 0b111

			rmText := getRM(rm, w, mod, file)
			regText := getRegister(reg, w)

			if d == 0b0 {
				fmt.Printf("%s %s, %s\n", command, rmText, regText)
			} else {
				fmt.Printf("%s %s, %s\n", command, regText, rmText)
			}
		} else if b1>>4 == 0b1011 {
			w := b1 >> 3 & 0b1
			reg := b1 & 0b111

			regText := getRegister(reg, w)

			if w == 0b1 {
				buf = make([]byte, 2)
			}

			_, err := file.Read(buf)
			if err == io.EOF {
				break
			}

			value := uint16(buf[0])
			if w == 0b1 {
				value = uint16(buf[1])<<8 | uint16(buf[0])
			}

			fmt.Printf("%s %s, %d\n", command, regText, value)
		} else if b1>>1 == 0b1100011 {
			w := b1 & 0b1

			buf := make([]byte, 1)
			_, err := file.Read(buf)
			if err == io.EOF {
				break
			}

			b2 := buf[0]

			mod := (b2 >> 6) & 0b11
			rm := b2 & 0b111

			rmText := getRM(rm, w, mod, file)

			if w == 0b1 {
				buf = make([]byte, 2)
			}

			_, err = file.Read(buf)
			if err == io.EOF {
				break
			}

			value := uint16(buf[0])
			if w == 0b1 {
				value = uint16(buf[1])<<8 | uint16(buf[0])
			}

			if w == 0b1 {
				fmt.Printf("%s %s, word %d\n", command, rmText, value)
			} else {
				fmt.Printf("%s %s, byte %d\n", command, rmText, value)
			}
		} else if b1>>1 == 0b1010000 || b1>>1 == 0b1010001 {
			w := b1 & 0b1
			p := b1 >> 1 & 0b1

			buf := make([]byte, 2)
			_, err := file.Read(buf)
			if err == io.EOF {
				break
			}

			value := uint16(buf[0])
			if w == 0b1 {
				value = uint16(buf[1])<<8 | uint16(buf[0])
			}

			if p == 0b1 {
				fmt.Printf("%s [%d], ax\n", command, value)
			} else {
				fmt.Printf("%s ax,[%d]\n", command, value)
			}
		}
	}
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

	bufSize := 1
	reg := regs[int(rm)]

	if mod == 0b00 {
		if rm == 0b110 {
			bufSize = 2
			reg = ""
		} else {
			bufSize = 0
		}
	}

	if mod == 0b10 {
		bufSize = 2
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
		return strings.Trim(fmt.Sprintf("[%s]", reg), " ")
	} else if reg == "" {
		return strings.Trim(fmt.Sprintf("[%d]", intValue), " ")
	}

	if intValue > 0 {
		return strings.Trim(fmt.Sprintf("[%s + %d]", reg, intValue), " ")
	} else {
		return strings.Trim(fmt.Sprintf("[%s - %d]", reg, -intValue), " ")
	}
}
