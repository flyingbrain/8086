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

	Register_a     registerIndex = "ax"
	Register_b     registerIndex = "bx"
	Register_c     registerIndex = "cx"
	Register_d     registerIndex = "dx"
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

var registerBuf [8]uint16

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Please set the file source")
	}

	name := args[1]
	run := false

	if args[1] == "exec" {
		run = true
		name = args[2]
	}

	file, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	buf := make([]byte, 2048)

	n, err := file.Read(buf)
	if err == io.EOF {
		fmt.Println("File fully read")
	}

	buf = buf[:n]
	var execBuf strings.Builder

	data, err := decode(buf)
	if run {
		exitCommand(data, &execBuf)
	}

	printCommand(data)
	fmt.Print(execBuf.String())

	if err != nil {
		fmt.Println(err)
	}
}

var regs = [8]registerIndex{
	Register_a,
	Register_b,
	Register_c,
	Register_d,
	Register_sp,
	Register_bp,
	Register_si,
	Register_di,
}

func exitCommand(data []decodedCommand, buf *strings.Builder) {
	fmt.Fprint(buf, "\n")
	for _, op := range data {
		switch op.optcode {
		case "mov":
			op.value[0].writeValue(op.value[1].getValue())
		}
	}

	for n, reg := range registerBuf {
		fmt.Fprintf(buf, "%s: 0x%04x (%d)\n", regs[n], reg, reg)
	}
}

func (o registerOperand) getValue() uint16 {
	value := registerBuf[o.reg]
	if o.size == 0 {
		if o.h == 0 {
			//low
			value = value & 0x00FF
		} else {
			//high
			value = (value >> 8) & 0xFF
		}
	}

	return value
}

func (o registerOperand) writeValue(val uint16) {
	regValue := o.getValue()
	if o.size == 0 {
		if o.h == 0 {
			//low
			regValue = (regValue & 0xFF00) | val
		} else {
			//high
			regValue = (regValue & 0x00FF) | (val << 8)
		}
	} else {
		regValue = val
	}

	registerBuf[o.reg] = regValue
}

func (o directOperand) getValue() uint16 {
	return uint16(o.value)
}
func (o directOperand) writeValue(val uint16) {
	fmt.Print(val)
}

func (o modOperand) getValue() uint16 {
	return uint16(o.value)
}
func (o modOperand) writeValue(val uint16) {
	fmt.Print(val)
}
