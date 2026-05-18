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

var zeroFlag bool
var signedFlag bool

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
	var strBuf strings.Builder
	fmt.Fprint(&strBuf, "bits 16\n\n")

	data, err := decode(buf)

	for _, com := range data {
		printCommand(com, &strBuf)

		if run {
			exitCommand(com, &strBuf)
		}

		fmt.Fprint(&strBuf, "\n")
	}

	if run {
		fmt.Fprint(&strBuf, "\n")
		fmt.Fprint(&strBuf, "Final registers:\n")
		for n, reg := range registerBuf {
			if reg != 0 {
				fmt.Fprintf(&strBuf, "    %s: 0x%04x (%d)\n", regs[n], reg, reg)
			}
		}
		flags := ""
		if zeroFlag {
			flags += "Z"
		}
		if signedFlag {
			flags += "S"
		}

		if len(flags) > 0 {
			fmt.Fprintf(&strBuf, "  flags: %s", flags)
		}
	}

	fmt.Print(strBuf.String())

	if err != nil {
		fmt.Println(err)
	}
}

func exitCommand(com decodedCommand, str *strings.Builder) {
	com.value[0].exec(com.optcode, com.value[1], str)
}

func (o registerOperand) writeValue(val uint16) {
	regValue := registerBuf[o.reg]
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

func (o registerOperand) movValue(s operand, str *strings.Builder) {
	value := registerBuf[o.reg]
	val := s.getValue()
	o.writeValue(val)

	fmt.Fprintf(str, " ; %s:0x%x->0x%x", regs[o.reg], value, val)
}

func (o registerOperand) subValue(s operand, str *strings.Builder) {
	value := registerBuf[o.reg]
	val := value - s.getValue()
	o.writeValue(val)

	fmt.Fprintf(str, " ; %s:0x%x->0x%x", regs[o.reg], value, val)
	checkFlags(val, str)
}

func (o registerOperand) cmpValue(s operand, str *strings.Builder) {
	value := registerBuf[o.reg]
	val := value - s.getValue()

	fmt.Fprintf(str, " ; %s:0x%x->0x%x", regs[o.reg], value, val)
	checkFlags(val, str)
}

func (o registerOperand) addValue(s operand, str *strings.Builder) {
	value := registerBuf[o.reg]
	val := value + s.getValue()
	o.writeValue(val)

	fmt.Fprintf(str, " ; %s:0x%x->0x%x", regs[o.reg], value, val)
	checkFlags(val, str)
}

func checkFlags(v uint16, str *strings.Builder) {
	on := ""
	off := ""
	if v == 0 && !zeroFlag {
		zeroFlag = true
		on += "Z"
	} else if v != 0 && zeroFlag {
		zeroFlag = false
		off += "Z"
	}

	isSigned := v&0x8000 == 0x8000

	if isSigned && !signedFlag {
		signedFlag = true
		on += "S"
	} else if !isSigned && signedFlag {
		signedFlag = false
		off += "S"
	}

	if len(on) > 0 || len(off) > 0 {
		fmt.Fprintf(str, " flags:%s->%s", off, on)
	}
}

func (o registerOperand) exec(optcode string, s operand, str *strings.Builder) {
	switch optcode {
	case "mov":
		o.movValue(s, str)
	case "sub":
		o.subValue(s, str)
	case "cmp":
		o.cmpValue(s, str)
	case "add":
		o.addValue(s, str)
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

func (o directOperand) exec(optcode string, s operand, str *strings.Builder) {
	log.Fatal("can not execute operation from direct value")
}
func (o directOperand) getValue() uint16 {
	return uint16(o.value)
}
func (o modOperand) exec(optcode string, s operand, str *strings.Builder) {
}
func (o modOperand) getValue() uint16 {
	return uint16(o.value)
}
