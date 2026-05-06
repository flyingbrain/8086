package main

import (
	"fmt"
)

type modOperand struct {
	base  effectiveAddressBase
	value int16
}

type directOperand struct {
	value int
	size  uint8
}

type register struct {
	reg  registerIndex
	h    int8
	size int8
}

type operand interface {
	printOp() string
}

type operation struct {
	value operand
}

type commandType string

const (
	jump  commandType = "jump"
	regul commandType = "regul"
)

type decodedCommand struct {
	value   [2]operation
	comType commandType
	optcode string
	amb     bool
	w       bool
}

func getReg(b byte, w int) operation {
	RegTable := [][2]register{
		{{Register_a, 0, 1}, {Register_a, 0, 2}},
		{{Register_c, 0, 1}, {Register_c, 0, 2}},
		{{Register_d, 0, 1}, {Register_d, 0, 2}},
		{{Register_b, 0, 1}, {Register_b, 0, 2}},
		{{Register_a, 1, 1}, {Register_sp, 0, 2}},
		{{Register_c, 1, 1}, {Register_bp, 0, 2}},
		{{Register_d, 1, 1}, {Register_si, 0, 2}},
		{{Register_b, 1, 1}, {Register_di, 0, 2}},
	}

	rez := operation{}
	rez.value = RegTable[int(b)][w]

	return rez
}

func getModName(b byte) effectiveAddressBase {

	tab := []effectiveAddressBase{
		EffectiveAddress_bx_si,
		EffectiveAddress_bx_di,
		EffectiveAddress_bp_si,
		EffectiveAddress_bp_di,
		EffectiveAddress_si,
		EffectiveAddress_di,
		EffectiveAddress_bp,
		EffectiveAddress_bx,
	}

	return tab[b]
}

func parceDispl(data []byte, pos *int, needed bool, w bool) int16 {
	if !needed {
		return 0
	}

	p := *pos / 8
	rez := int16(data[p])
	*pos += 8

	if w {
		if data[p+1] != 0 {
			rez = int16(data[p+1])<<8 | int16(rez)
		}
		*pos += 8
	}

	return int16(rez)
}

func decode(buf []byte) ([]decodedCommand, error) {
	cmds := []decodedCommand{}

	for len(buf) > 0 {
		matched := false
		pos := 0

		for _, command := range Commands {
			if readCommand(buf, command, &cmds, &pos) {
				matched = true
				buf = buf[pos/8:]
				break
			}
		}

		if !matched {
			fmt.Printf("can not transalate instruction %08b, trying next byte\n", buf[0])
			buf = buf[1:]
		}

	}
	return cmds, nil
}

func readCommand(buf []byte, com Command, cmds *[]decodedCommand, pos *int) bool {
	f := map[DataType]byte{}

	for _, c := range com.value {

		n := 0
		p := 0
		if *pos != 0 {
			n = *pos / 8
			p = *pos % 8
		}

		b := buf[n]
		if c.Name == Literal && (b<<p)>>(8-int(c.Width)) != c.Value {
			*pos = 0
			return false
		}

		if c.HasValue {
			f[c.Name] = c.Value
		} else {
			f[c.Name] = (b << p) >> (8 - int(c.Width))
		}

		*pos += int(c.Width)
	}

	if *pos%8 != 0 {
		panic(fmt.Sprintf("Commad decode wrong position:%d", *pos))
	}

	command := decodedCommand{
		value:   [2]operation{},
		comType: regul,
		optcode: com.Name,
		amb:     true,
		w:       false,
	}

	D := f[dType]
	W := f[wType]
	S, isS := f[sType]
	MOD := f[modType]
	rm := f[rmType]
	_, isDisp := f[dispType]

	directAddr := MOD == 0b00 && rm == 0b110
	hasDispl := MOD == 0b10 || MOD == 0b01 || directAddr || isDisp
	DisplW := directAddr || MOD == 0b10 || W != 0 && isDisp

	dispData := parceDispl(buf, pos, hasDispl, DisplW)

	_, isData := f[dataType]
	hasData := isData || isS
	dataW := isS && S == 1 || W == 0

	command.w = DisplW || !dataW

	// REG
	if _, ok := f[regType]; ok {
		command.value[boolToInt(D == 0)] = getReg(f[regType], int(W))
		command.amb = false
	}

	// RM
	if _, ok := f[rmType]; ok {
		if MOD == 0b11 {
			command.value[D] = getReg(f[rmType], int(W))
			command.amb = false
		} else {

			modName := getModName(f[rmType])
			if directAddr {
				modName = EffectiveAddress_direct
			}

			operand := modOperand{
				base:  modName,
				value: dispData,
			}

			command.value[D] = operation{
				value: operand,
			}
		}
	} else if dispData != 0 {
		operand := modOperand{
			base:  EffectiveAddress_direct,
			value: dispData,
		}

		command.comType = jump

		command.value[1] = operation{
			value: operand,
		}
	}

	// direct data
	if hasData {
		data := parceDispl(buf, pos, hasData, !dataW)
		dataOper := directOperand{
			value: int(data),
			size:  uint8(W),
		}

		op := operation{}

		idx := 0
		if command.value[0] != op {
			idx = 1
		}

		op.value = dataOper

		command.value[idx] = op
	}

	*cmds = append(*cmds, command)
	return true
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
