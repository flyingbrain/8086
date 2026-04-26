package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type registerIndex string

const (
	Register_none registerIndex = ""

	Register_a     registerIndex = "a"
	Register_b     registerIndex = "b"
	Register_c     registerIndex = "c"
	Register_d     registerIndex = "d"
	Register_sp    registerIndex = "sp"
	Register_bp    registerIndex = "bp"
	Register_si    registerIndex = "si"
	Register_di    registerIndex = "di"
	Register_es    registerIndex = "es"
	Register_cs    registerIndex = "cs"
	Register_ss    registerIndex = "ss"
	Register_ds    registerIndex = "ds"
	Register_ip    registerIndex = "ip"
	Register_flags registerIndex = "flags"

	Register_count registerIndex = "count"
)

type effectiveAddressBase string

const (
	EffectiveAddress_direct effectiveAddressBase = ""

	EffectiveAddress_bx_si effectiveAddressBase = "bx + si"
	EffectiveAddress_bx_di effectiveAddressBase = "bx + di"
	EffectiveAddress_bp_si effectiveAddressBase = "bp + si"
	EffectiveAddress_bp_di effectiveAddressBase = "bp + di"
	EffectiveAddress_si    effectiveAddressBase = "si"
	EffectiveAddress_di    effectiveAddressBase = "di"
	EffectiveAddress_bp    effectiveAddressBase = "bp"
	EffectiveAddress_bx    effectiveAddressBase = "bx"

	EffectiveAddress_count effectiveAddressBase = "count"
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

	buf := make([]byte, 1024)

	n, err := file.Read(buf)
	if err == io.EOF {

	}
	buf = buf[:n]

	data, err := decode(buf)
	printCommand(data)

	if err != nil {
		fmt.Println(err)
	}
}

func printCommand(data []decodedCommand) {
	fmt.Print("bits 16\n\n")
	var str strings.Builder

	for _, com := range data {

		fmt.Fprintf(&str, "%s ", com.optcode)
		sep := ", "

		for _, c := range com.value {
			fmt.Fprintf(&str, "%s%s", c.value.printO(), sep)
			sep = ""
		}

		str.WriteString("\n")
	}

	fmt.Print(str.String())
}

func (o modOperand) printO() string {
	if o.value != 0 {
		return fmt.Sprintf("[%s + %d]", o.base, o.value)
	}

	return fmt.Sprintf("%s", o.base)
}

func (o register) printO() string {
	c := ""
	if o.size == 2 && len(o.reg) == 1 {
		c = "x"
	} else if o.size == 1 {
		if o.h == 1 {
			c = "h"
		} else {
			c = "l"
		}
	}

	return string(o.reg) + c
}
